package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/stianeikeland/go-rpio"
)

var (
	_buildDate    string
	_buildVersion string
	log           = logrus.New()
	_relays       []rpio.Pin
	mqttClient    mqtt.Client
)

func main() {
	var err error
	log.SetLevel(logrus.TraceLevel)
	log.Printf("---------- Program Started %v (%v) ----------", _buildVersion, _buildDate)

	err = setupRelays(os.Getenv("RELAYS"))
	if err != nil {
		log.Panic(err)
	}
	defer rpio.Close()

	connectToMQTT()

	mqttClient.Publish("testChannel/Topic", 0x01, false, "testMessage")

	select {}
}

func connectToMQTT() error {
	ServerAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:10001")
	if err != nil {
		return err
	}
	LocalAddr, err := net.ResolveUDPAddr("udp", ":10002")
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", LocalAddr)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	buf := make([]byte, 1024)

	n, err := udpConn.WriteTo([]byte{0x01}, ServerAddr)
	if err != nil {
		return err
	}
	log.Tracef("packet-written: bytes=%d to=%s\n", n, ServerAddr.String())

	n, addr, err := udpConn.ReadFromUDP(buf)
	log.Trace("Received ", string(buf[0:n]), " from ", addr)
	if err != nil {
		return err
	}

	BrokerAddr := fmt.Sprintf("%v%v", addr.IP, ":1883")
	log.Trace("Broker ", BrokerAddr)
	mqttClientOptions := mqtt.NewClientOptions()
	mqttClientOptions.AddBroker(BrokerAddr)
	mqttClient = mqtt.NewClient(mqttClientOptions)
	token := mqttClient.Connect()
	for token.Wait() && token.Error() != nil && mqttClient.IsConnected() == false {
		time.Sleep(2 * time.Second)
		log.Error("Trying to Connect to Broker", BrokerAddr)
		token = mqttClient.Connect()
	}
	log.Info("Connected to Broker:", BrokerAddr)
	return nil
}

func setupRelays(RelaysString string) error {
	var err error
	log.Trace("ENV 'RELAYS'=", RelaysString)
	RelaysStringArray := strings.Split(RelaysString, ",")

	log.Trace("Opening GPIO")
	err = rpio.Open()
	if err != nil {
		return err
	}
	log.Trace("Opened GPIO Successfully")

	log.Trace("Setting Up GPIO")
	for _, RelayString := range RelaysStringArray {
		RelayPosition, err := strconv.Atoi(RelayString)
		if err != nil {
			log.Error(err)
			continue
		}
		Relay := rpio.Pin(RelayPosition)
		Relay.Output()
		Relay.High()
		_relays = append(_relays, Relay)
	}
	log.Trace("Setup GPIO Successfully")
	return nil
}
