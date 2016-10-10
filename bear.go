//author 逆雪寒
//version 0.6.1
//分布式id生成服务
package main

import(
    "net"
    "fmt"
    "log"
    "flag"
    "io"
    "time"
    "bytes"
    "sync"
    "encoding/binary"
    "math/rand"
    "runtime"
)

var host = flag.String("h","localhost","Bound IP. default:localhost")
var port = flag.String("p","8384","port. default:8384")
var room = flag.Int("r",1,"server room. default:1")
var serverId = flag.Int("s",1,"server id. default:1")

type Fish struct {
    millis int64
    increase int64
}
var Pool map[uint8]Fish

type CommandCode uint8
type Status uint8

const (
    REQ_MAGIC = 0x83
    RES_MAGIC = 0x84
    B_LEN = 4
)
const (
    GET = CommandCode(0x01)
)
const (
    SUCCESS = Status(0x33)
    FAIL    = Status(0x44)
)

var handlers = map[CommandCode]func(Request) Response{
    GET : GetIdHandler,
}

func init() {
    Pool = make(map[uint8]Fish)
}

func must(e error) {
    if e != nil {
        panic(e)
    }
}

func random() int{
    x := rand.Intn(9)
    if x == 0 {
        x = 1
    }
    return x
}

func ERROR(w io.WriteCloser){
    defer func(){must(w.Close())}()

    var body = make([]byte,2)
    body[0] = RES_MAGIC
    body[1] = byte(FAIL)
    w.Write(body)
}

type Response struct {
    Body []byte
}

func(this *Response)HeaderBodyFill(status Status,data interface{}) []byte{
    var body = make([]byte,2)
    body[0] = RES_MAGIC
    body[1] = byte(SUCCESS)

    buf := bytes.NewBuffer([]byte{}) 
    binary.Write(buf, binary.BigEndian,data)
    body = append(body,buf.Bytes()...)
    return body
}

func(this *Response)Write(w io.WriteCloser) (int,error){
    defer func(){must(w.Close())}()
    return w.Write(this.Body)
}

func GetIdHandler(req Request) Response{
    now := time.Now()
    nanos := now.UnixNano()
    millis := nanos / 1000000

    workId := req.WordId
    //毫秒41 + 机房2 + 机器5 + 业务8+ 序列号7
    id := (millis << 22) + (int64(*room) << 20) + (int64(*serverId) << 15) + (int64(workId) << 7)
    
    var uniqueID int64

    req.Dog.Lock()
    if v,ok := Pool[workId]; ok {
        if v.millis ==  millis {
            v.increase += 1
        }else{
            v.millis = millis
            v.increase = int64(random())
        }
        uniqueID = id + v.increase
    }else{
        Pool[workId] = Fish{millis,1}
        uniqueID = id + 1
    }
    req.Dog.Unlock()

    res := Response{}
    res.Body = res.HeaderBodyFill(SUCCESS,uniqueID)

    return res
}

func flying(reqChan chan Request) {
    for{
        req := <- reqChan
        if kfc,ok := handlers[req.CommandCode]; ok {
            res := kfc(req)
            res.Write(req.Conn)
        }else{
            log.Printf("commandCode error:%v",req.CommandCode)
            ERROR(req.Conn)
        }
    }
}

type Request struct {
    CommandCode CommandCode
    WordId uint8
    Conn net.Conn
    Dog *sync.Mutex
}

func IO(conn net.Conn,msgBytes []byte) (Request,error){
    if msgBytes[0] != REQ_MAGIC {
        ERROR(conn)
        return Request{},fmt.Errorf("bad 0x%x",msgBytes[0])
    }

    req := Request{}
    req.CommandCode = CommandCode(msgBytes[1])
    req.WordId = uint8(msgBytes[2]) //业务id
    return req,nil
}

func handler(r io.Reader,conn net.Conn,reqChan chan Request) error{
    buf := make([]byte,B_LEN)
    if _,e := io.ReadFull(r,buf);e != nil {
        return e
    }

    lock := &sync.Mutex{}

    if req,e := IO(conn,buf);e == nil {
        req.Conn = conn
        req.Dog = lock
        reqChan <- req
    }else{
        log.Printf("Format error:%s",e)
    }
    return io.EOF
}

func waitForYou(ls net.Listener) {
    var reqChan = make(chan Request)

    go flying(reqChan)

    for {
        conn,e := ls.Accept()
        if e == nil {
            go handler(conn,conn,reqChan)
        }else{
            log.Printf("Error accepting from %s",ls)
        }
    }
}

func main() {
    defer func() {
        if err := recover();err != nil{
            log.Printf("Fatal error:%v",err)
        }
    }()

    runtime.GOMAXPROCS(runtime.NumCPU())

    flag.Parse()
    log.SetFlags(log.LstdFlags)
    ls,e := net.Listen("tcp",*host + ":" + *port)
    if e != nil {
        log.Fatalf("Error bound: %s",e)
    }
    waitForYou(ls)
}