<?php
namespace Bad;
/**
 * author 逆雪寒
 * version 0.6.1
 */
class Bear {
    const RES_MAGIC = 0x84;
    const GET_ID_COMMAMD = 0x01;
    const SUCCESS = 0x33;

	private static $instance = NULL;
	private $host = '';
    private $port = 0;

	private function __construct($host,$port,$time){ 
		$this->host = $host;
        $this->port = $port;
        $this->timeout = $time;
	}
	 
	public function __clone(){
		trigger_error('Clone is not allow!',E_USER_ERROR);
	}

    private function ordering($command,$workID) {
        $fp=fsockopen($this->host,$this->port,$errno,$err,$this->timeout); 
        $data = '';
        if($fp){ 
            fwrite($fp,pack("N",(hexdec("0x83") << 24) + (hexdec($command) << 16) + ($workID << 8))); 
            stream_set_blocking($fp,true);
            stream_set_timeout($fp,$this->timeout);
            $info=stream_get_meta_data($fp);
            while((!feof($fp)) && (!$info['timed_out'])){ 
                $data.=fgets($fp,11); 
            }
        }
        return $data;
    }

    private function getBody($dataBytes) {
        $size = strlen($dataBytes) - 2;
        $data = unpack("Cmagic/Cstatus/C$size",$dataBytes);
        if($data['magic'] != self::RES_MAGIC || $data['status'] != self::SUCCESS) {
            return false;
        }
        $data = array_slice($data,2);
        return $data;
    }

	public function factory($host = 'localhost',$port = 8384,$time = 1) {
		if(is_null(self::$instance)) {
			self::$instance = new self($host,$port,$time);
		}
		return self::$instance;
	}

	/**
	 * 获取全局id.
	 * @param int8 $workID  业务id  范围 1 - 255   比如 1:用户  2:动态 又或者  1:微博系统  2:招聘系统
	 * @return bool or int64
	 */
    public function ID($workID = 1) {
        if($workID < 1 || $workID > 255) {
            return false;
        }

        $dataBytes = $this->ordering(self::GET_ID_COMMAMD,$workID);
        $bodyBytes = $this->getBody($dataBytes);

        if($bodyBytes === false) {
            return false;
        }

        for($id=0,$i = 56,$j=0;$i >= 0;$i = $i - 8,$j++) {
            $id += $bodyBytes[$j] << $i;
        }
        return $id;
    }

    /**
     * 获取业务id.
     * @param int64 $id
     * @return bool or int8
     */
    public function workID($id) {
        $workID = ($id >> 7) & 0xFF;
        return $workID;
    }

    /**
     * 获取ID生成时间戳
     * @param int64 $id
     * @return int
     */
    public function time($id) {
        $millis = (($id >> 22) & 0x1FFFFFFFFFF) / 1000;
        return (int)$millis;
    }
}

// //example
// $bear = Bear::factory("127.0.0.1");
// $id = $bear->ID(233);
// $a = $bear->time($id);


