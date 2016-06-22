bear-分布式ID生成服务
============

sdk 目前只有 php 版本:
https://github.com/nixuehan/bear/tree/master/sdk


按照这个格式生成全局唯一id： 毫秒40bit + 机房2bit + 机器6bit + 业务8bit + 序列号7bit 

下载符合自己机器的版本，运行即可


    $ ./bear


###支持的参数：

  -h string
    	Bound IP. default:localhost (default "localhost")

  -p string
    	port. default:8384 (default "8384")

  -r int
    	server room. default:1 (default 1)

  -s int
    	server id. default:1 (default 1)




###高可用建议：

 haprxoy	
 			->  bear1 -s 1

 			->  bear2 -s 2


 其他不懂 看源代码吧