package sconsul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/syncfuture/go/sconfig"
	"github.com/syncfuture/go/u"
	"github.com/syncfuture/host/service"
)

func RegisterServiceInfo(cp sconfig.IConfigProvider, host service.IServiceHost) {
	// 读取配置
	consulAddr := cp.GetString("Consul.Addr")
	serviceName := cp.GetString("Consul.Service.Name")
	serviceCheckTimeout := cp.GetString("Consul.Service.Check.Timeout")
	serviceCheckInterval := cp.GetString("Consul.Service.Check.Interval")
	serviceHost := cp.GetString("Consul.Service.Host")
	servicePort := cp.GetInt("Consul.Service.Port")

	// 服务中心客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr
	consulClient, err := api.NewClient(consulConfig)
	u.LogFaltal(err)
	consulAgent := consulClient.Agent()

	// 在服务中心登记服务
	err = consulAgent.ServiceRegister(&api.AgentServiceRegistration{
		ID:   fmt.Sprintf("%v[%v:%v]", serviceName, host.GetHost(), host.GetPort()), // 服务节点的名称
		Name: serviceName,                                                           // 服务名称
		// Tags:    r.Tag,                                        // tag，可以为空
		Address: serviceHost, // 服务 IP
		Port:    servicePort, // 服务端口
		Check: &api.AgentServiceCheck{ // 健康检查
			Interval:                       serviceCheckInterval, // 健康检查间隔
			TCP:                            host.GetListenAddr(),
			DeregisterCriticalServiceAfter: serviceCheckTimeout, // 注销时间，相当于过期时间
		},
	})
	u.LogFaltal(err)
}
