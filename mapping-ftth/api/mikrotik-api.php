<?php
header('Content-Type: application/json');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST');
header('Access-Control-Allow-Headers: Content-Type');

// Mikrotik RouterOS API Class
class MikrotikAPI {
    private $host;
    private $user;
    private $pass;
    private $port;
    private $socket;
    
    public function __construct($host, $user, $pass, $port = 8728) {
        $this->host = $host;
        $this->user = $user;
        $this->pass = $pass;
        $this->port = $port;
    }
    
    public function connect() {
        $this->socket = @fsockopen($this->host, $this->port, $errno, $errstr, 3);
        
        if (!$this->socket) {
            return ['success' => false, 'error' => "Connection failed: $errstr ($errno)"];
        }
        
        // Login
        $this->write('/login');
        $response = $this->read();
        
        if (isset($response[0]['!done'])) {
            $this->write('/login', false, [
                '=name=' . $this->user,
                '=password=' . $this->pass
            ]);
            
            $loginResponse = $this->read();
            
            if (isset($loginResponse[0]['!done'])) {
                return ['success' => true];
            }
        }
        
        return ['success' => false, 'error' => 'Login failed'];
    }
    
    public function getActivePPPoE() {
        if (!$this->socket) {
            return ['success' => false, 'error' => 'Not connected'];
        }
        
        $this->write('/ppp/active/print');
        $response = $this->read();
        
        $activeUsers = [];
        
        foreach ($response as $item) {
            if (isset($item['!re'])) {
                $activeUsers[] = [
                    'name' => isset($item['=name']) ? $item['=name'] : '',
                    'address' => isset($item['=address']) ? $item['=address'] : '',
                    'uptime' => isset($item['=uptime']) ? $item['=uptime'] : '',
                    'caller-id' => isset($item['=caller-id']) ? $item['=caller-id'] : '',
                    'service' => isset($item['=service']) ? $item['=service'] : ''
                ];
            }
        }
        
        return ['success' => true, 'data' => $activeUsers];
    }
    
    private function write($command, $param2 = false, $param3 = []) {
        $this->writeWord(strlen($command));
        fwrite($this->socket, $command);
        
        if ($param2) {
            $this->writeWord(strlen($param2));
            fwrite($this->socket, $param2);
        }
        
        foreach ($param3 as $attr) {
            $this->writeWord(strlen($attr));
            fwrite($this->socket, $attr);
        }
        
        fwrite($this->socket, chr(0));
    }
    
    private function writeWord($len) {
        if ($len < 0x80) {
            fwrite($this->socket, chr($len));
        } elseif ($len < 0x4000) {
            $len |= 0x8000;
            fwrite($this->socket, chr(($len >> 8) & 0xFF));
            fwrite($this->socket, chr($len & 0xFF));
        } elseif ($len < 0x200000) {
            $len |= 0xC00000;
            fwrite($this->socket, chr(($len >> 16) & 0xFF));
            fwrite($this->socket, chr(($len >> 8) & 0xFF));
            fwrite($this->socket, chr($len & 0xFF));
        } elseif ($len < 0x10000000) {
            $len |= 0xE0000000;
            fwrite($this->socket, chr(($len >> 24) & 0xFF));
            fwrite($this->socket, chr(($len >> 16) & 0xFF));
            fwrite($this->socket, chr(($len >> 8) & 0xFF));
            fwrite($this->socket, chr($len & 0xFF));
        } else {
            fwrite($this->socket, chr(0xF0));
            fwrite($this->socket, chr(($len >> 24) & 0xFF));
            fwrite($this->socket, chr(($len >> 16) & 0xFF));
            fwrite($this->socket, chr(($len >> 8) & 0xFF));
            fwrite($this->socket, chr($len & 0xFF));
        }
    }
    
    private function read() {
        $response = [];
        
        while (true) {
            $word = $this->readWord();
            
            if ($word === false || strlen($word) == 0) {
                break;
            }
            
            $response[] = $this->parseResponse($word);
            
            if (strpos($word, '!done') === 0) {
                break;
            }
        }
        
        return $response;
    }
    
    private function readWord() {
        $len = $this->readLen();
        
        if ($len === false || $len == 0) {
            return '';
        }
        
        $word = '';
        $word = fread($this->socket, $len);
        
        return $word;
    }
    
    private function readLen() {
        $byte = ord(fread($this->socket, 1));
        
        if ($byte == 0) {
            return 0;
        }
        
        if (($byte & 0x80) == 0x00) {
            return $byte;
        }
        
        if (($byte & 0xC0) == 0x80) {
            return (($byte & ~0xC0) << 8) + ord(fread($this->socket, 1));
        }
        
        if (($byte & 0xE0) == 0xC0) {
            $len = (($byte & ~0xE0) << 8) + ord(fread($this->socket, 1));
            return ($len << 8) + ord(fread($this->socket, 1));
        }
        
        if (($byte & 0xF0) == 0xE0) {
            $len = (($byte & ~0xF0) << 8) + ord(fread($this->socket, 1));
            $len = ($len << 8) + ord(fread($this->socket, 1));
            return ($len << 8) + ord(fread($this->socket, 1));
        }
        
        if (($byte & 0xF8) == 0xF0) {
            $len = ord(fread($this->socket, 1));
            $len = ($len << 8) + ord(fread($this->socket, 1));
            $len = ($len << 8) + ord(fread($this->socket, 1));
            return ($len << 8) + ord(fread($this->socket, 1));
        }
        
        return false;
    }
    
    private function parseResponse($response) {
        $parsed = [];
        $lines = explode("\n", $response);
        
        foreach ($lines as $line) {
            if (strpos($line, '=') !== false) {
                list($key, $value) = explode('=', $line, 2);
                $parsed['=' . $key] = $value;
            } elseif (strpos($line, '!') === 0) {
                $parsed[$line] = true;
            }
        }
        
        if (empty($parsed)) {
            $parsed[$response] = true;
        }
        
        return $parsed;
    }
    
    public function disconnect() {
        if ($this->socket) {
            fclose($this->socket);
        }
    }
}

// Load configuration
$configFile = __DIR__ . '/config.php';
if (file_exists($configFile)) {
    include $configFile;
} else {
    // Default config (will be overridden by config.php)
    $MIKROTIK_HOST = '192.168.88.1';
    $MIKROTIK_USER = 'admin';
    $MIKROTIK_PASS = '';
    $MIKROTIK_PORT = 8728;
}

// Handle request
$action = isset($_GET['action']) ? $_GET['action'] : 'status';

switch ($action) {
    case 'status':
        // Get active PPPoE connections
        $mikrotik = new MikrotikAPI($MIKROTIK_HOST, $MIKROTIK_USER, $MIKROTIK_PASS, $MIKROTIK_PORT);
        $connectResult = $mikrotik->connect();
        
        if (!$connectResult['success']) {
            echo json_encode([
                'success' => false,
                'error' => $connectResult['error']
            ]);
            break;
        }
        
        $result = $mikrotik->getActivePPPoE();
        $mikrotik->disconnect();
        
        echo json_encode($result);
        break;
        
    case 'test':
        // Test connection
        $mikrotik = new MikrotikAPI($MIKROTIK_HOST, $MIKROTIK_USER, $MIKROTIK_PASS, $MIKROTIK_PORT);
        $connectResult = $mikrotik->connect();
        
        if ($connectResult['success']) {
            $mikrotik->disconnect();
            echo json_encode([
                'success' => true,
                'message' => 'Connection successful'
            ]);
        } else {
            echo json_encode($connectResult);
        }
        break;
        
    default:
        echo json_encode([
            'success' => false,
            'error' => 'Invalid action'
        ]);
}
?>
