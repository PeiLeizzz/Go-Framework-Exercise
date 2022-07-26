package main

import (
	"encoding/json"
	"errors"
	"go-product/RabbitMQ"
	"go-product/imooc-product/common"
	"go-product/imooc-product/datamodels"
	"go-product/imooc-product/encrypt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

func Auth(w http.ResponseWriter, req *http.Request) error {
	err := checkUserInfo(req)
	if err != nil {
		return err
	}
	return nil
}

func CheckRight(w http.ResponseWriter, req *http.Request) {
	right := accessControl.GetDistributedRight(req)
	if !right {
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
	return
}

func Check(w http.ResponseWriter, req *http.Request) {
	queryForm, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 || len(queryForm["productID"][0]) <= 0 {
		w.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	userCookie, err := req.Cookie("uid")
	if err != nil {
		w.Write([]byte("false"))
		return
	}

	// 分布式权限验证
	right := accessControl.GetDistributedRight(req)
	if right == false {
		w.Write([]byte("false"))
		return
	}

	// 获取数量控制权限
	hostUrl := "http://" + "127.0.0.1:8084/getOne"
	responseValidate, validateBody, err := GetCurl(hostUrl, req)
	if err != nil {
		w.Write([]byte("false"))
		return
	}

	// 判断数量控制接口请求状态
	if responseValidate.StatusCode == 200 {
		if string(validateBody) == "true" {
			productID, err := strconv.ParseInt(productString, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			userID, err := strconv.ParseInt(userCookie.Value, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			message := datamodels.NewMessage(productID, userID)
			byteMessage, err := json.Marshal(message)
			if err != nil {
				w.Write([]byte("false"))
				return
			}

			rabbitMqValidate.Publish(string(byteMessage))
			w.Write([]byte("true"))
			return
		}
	}

	w.Write([]byte("false"))
}

// 身份校验函数
func checkUserInfo(req *http.Request) error {
	// 获取 cookie
	uidCookie, err := req.Cookie("uid")
	if err != nil {
		return err
	}
	SignCookie, err := req.Cookie("sign")
	if err != nil {
		return err
	}

	signByte, err := encrypt.DePwdCode(SignCookie.Value)
	if err != nil {
		return errors.New("加密串被篡改")
	}

	if !checkInfo(uidCookie.Value, string(signByte)) {
		return errors.New("身份认证未通过")
	}

	return nil
}

func checkInfo(checkStr string, signStr string) bool {
	if checkStr == "" || signStr == "" || checkStr != signStr {
		return false
	}
	return true
}

var (
	// 集群地址
	hosts = []string{"127.0.0.1", "127.0.0.1"}
	// 本机地址
	localhost       = "127.0.0.1"
	port            = "8083"
	hashConsisitent *common.Consistent
)

// 用来存放控制信息
type AccessControl struct {
	// 存放用户想要存放的信息
	sources map[int64]interface{}
	mu      sync.RWMutex
}

var accessControl = &AccessControl{sources: make(map[int64]interface{})}

// 获取指定数据
func (m *AccessControl) GetNewRecord(uid int64) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data := m.sources[uid]
	return data
}

// 设置数据
func (m *AccessControl) SetNewRecord(uid int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sources[uid] = "hello imooc"
}

// 获取本机 map 并处理业务逻辑
func (m *AccessControl) GetDataFromMap(uid string) bool {
	//uidInt64, err := strconv.ParseInt(uid, 10, 64)
	//if err != nil {
	//	return false
	//}
	//
	//data := m.GetNewRecord(uidInt64)
	//
	//return data != nil
	return true
}

// 模拟请求
func GetCurl(hostUrl string, req *http.Request) (response *http.Response, body []byte, err error) {
	uidPre, err := req.Cookie("uid")
	if err != nil {
		return
	}

	uidSign, err := req.Cookie("sign")
	if err != nil {
		return
	}

	// 模拟接口访问
	client := &http.Client{}
	req, err = http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		return
	}

	cookieUid := &http.Cookie{Name: "uid", Value: uidPre.Value, Path: "/"}
	cookieSign := &http.Cookie{Name: "sign", Value: uidSign.Value, Path: "/"}
	req.AddCookie(cookieUid)
	req.AddCookie(cookieSign)

	response, err = client.Do(req)
	defer response.Body.Close()
	if err != nil {
		return
	}

	body, err = ioutil.ReadAll(response.Body)
	return
}

// 获取其他节点的处理结果
func (m *AccessControl) GetDataFromOtherHost(host string, req *http.Request) bool {
	hostUrl := "http://" + host + ":" + port + "/checkRight"
	response, body, err := GetCurl(hostUrl, req)
	if err != nil {
		return false
	}

	if response.StatusCode == 200 {
		return string(body) == "true"
	}
	return false
}

func (m *AccessControl) GetDistributedRight(req *http.Request) bool {
	uid, err := req.Cookie("uid")
	if err != nil {
		return false
	}

	hostRequest, err := hashConsisitent.Get(uid.Value)
	if err != nil {
		return false
	}

	// 判断是否为本机
	if hostRequest == localhost {
		// 执行本机数据读取和校验
		return m.GetDataFromMap(uid.Value)
	} else {
		// 不是本机，则充当代理访问数据，返回结果
		return m.GetDataFromOtherHost(hostRequest, req)
	}
}

var rabbitMqValidate *RabbitMQ.RabbitMQSimple

func main() {
	// 负载均衡器设置

	// 采用一致性哈希算法
	hashConsisitent = common.NewConsistent()
	for _, v := range hosts {
		hashConsisitent.Add(v)
	}

	// 自动获取 ip
	localIP, err := common.GetIntranceIP()
	if err != nil {
		log.Fatal(err)
	}
	localhost = localIP

	rabbitMqValidate = RabbitMQ.NewRabbitMQSimple("imoocProduct")
	defer rabbitMqValidate.Destroy()

	// 1. 启动过滤器
	filter := common.NewFilter()

	// 2. 注册拦截器
	filter.RegisterFilterUri("/check", Auth)
	filter.RegisterFilterUri("/checkRight", Auth)

	http.HandleFunc("/check", filter.Handle(Check))
	http.HandleFunc("/checkRight", filter.Handle(CheckRight))
	http.ListenAndServe(":8083", nil)
}
