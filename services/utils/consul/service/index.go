package consulservice

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/hashicorp/consul/api"
)

type Service struct {
	*api.Client
	config       *Config
	registration *api.AgentServiceRegistration
	registered   bool
	listener     net.Listener
	closeChan    chan struct{}
	wg           sync.WaitGroup
}

type Config struct {
	ServiceId      string
	ServiceName    string
	ServicePort    string
	ServiceAddress string
}

func New(client *api.Client, config *Config) (*Service, error) {

	tcpAddr := fmt.Sprintf("%s:%v", config.ServiceAddress, config.ServicePort)

	port, err := strconv.Atoi(config.ServicePort)
	if err != nil {
		return nil, err
	}

	registration := &api.AgentServiceRegistration{
		ID:      config.ServiceId,
		Name:    config.ServiceName,
		Address: config.ServiceAddress,
		Port:    port,
		Check: &api.AgentServiceCheck{
			TCP:                            tcpAddr,
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return &Service{
		Client:       client,
		config:       config,
		registration: registration,
	}, nil
}

func (s *Service) Register() error {
	err := s.Agent().ServiceRegister(s.registration)
	if err != nil {
		return err
	}
	s.registered = true
	return nil
}

func (s *Service) Deregister() error {
	err := s.Agent().ServiceDeregister(s.registration.ID)
	if err != nil {
		return err
	}
	s.registered = false
	return nil
}

func (s *Service) Close() {
	log.Println("Closing service...")

	if s.registered {
		s.Deregister()
	}

	if s.closeChan != nil {
		close(s.closeChan)
	}

	s.wg.Wait()
	log.Println("Service deregistered and closed.")
}

func (s *Service) RunTCP(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	closeChan := make(chan struct{})
	s.closeChan = closeChan

	log.Println("TCP server started on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-closeChan:
				log.Println("Shutting down TCP listener...")
				return nil
			default:
				log.Println("Error accepting connection:", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Service) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()

	response := "alive"
	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Println("Error writing to connection:", err)
		return
	}
}

func GetLocalIPV4Address() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no reachable IPv4 address found")
}
