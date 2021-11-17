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
	consulAddr := cp.GetString("ConsulAddr")
	serviceName := cp.GetString("ServiceName")
	serviceTimeout := cp.GetString("ServiceTimeout")
	checkInterval := cp.GetString("CheckInterval")

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
		Port:    host.GetPort(), // 服务端口
		Address: host.GetHost(), // 服务 IP
		Check: &api.AgentServiceCheck{ // 健康检查
			Interval:                       checkInterval, // 健康检查间隔
			TCP:                            host.GetListenAddr(),
			DeregisterCriticalServiceAfter: serviceTimeout, // 注销时间，相当于过期时间
		},
	})
	u.LogFaltal(err)
}
