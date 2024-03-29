package sconsul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/u"
)

func RegisterServiceInfo(cp sconfig.IConfigProvider) {
	// 读取配置
	consulAddr := cp.GetString("Consul.Addr")
	consulToken := cp.GetString("Consul.Token")
	serviceName := cp.GetString("Consul.Service.Name")
	serviceCheckTimeout := cp.GetString("Consul.Service.Check.Timeout")
	serviceCheckInterval := cp.GetString("Consul.Service.Check.Interval")
	serviceHost := cp.GetString("Consul.Service.Host")
	servicePort := cp.GetInt("Consul.Service.Port")
	serviceID := fmt.Sprintf("%v:%v", serviceHost, servicePort)

	// 服务中心客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr
	consulConfig.Token = consulToken
	consulClient, err := api.NewClient(consulConfig)
	u.LogFatal(err)
	consulAgent := consulClient.Agent()

	// 在服务中心登记服务
	err = consulAgent.ServiceRegister(&api.AgentServiceRegistration{
		ID:   serviceID,   // 服务节点的名称
		Name: serviceName, // 服务名称
		// Tags:    r.Tag,                                        // tag，可以为空
		Address: serviceHost, // 服务 IP
		Port:    servicePort, // 服务端口
		Check: &api.AgentServiceCheck{ // 健康检查
			Interval:                       serviceCheckInterval, // 健康检查间隔
			TCP:                            fmt.Sprintf("%s:%d", serviceHost, servicePort),
			DeregisterCriticalServiceAfter: serviceCheckTimeout, // 注销时间，相当于过期时间
		},
	})
	u.LogFatal(err)
}
