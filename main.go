package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2868"
	"layeh.com/radius/rfc3580"
)

type Server struct {
	Listen      string
	Secret      string
	DefaultVlan int
}

type Vlan struct {
	Name string
	Id   int
}

type Client struct {
	Name string
	Vlan string
	Mac  string
}

type Config struct {
	Server  Server
	Clients []Client
	Vlans   []Vlan
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.json", "The first number! Default is 1")
	flag.Parse()

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	var vlanMap = make(map[string]string)
	for _, vlan := range config.Vlans {
		vlanMap[vlan.Name] = strconv.Itoa(vlan.Id)
	}

	var clientMap = make(map[string]Client)
	for _, client := range config.Clients {
		clientMap[normalizedMac(client.Mac)] = client
	}

	var defaultVlan = strconv.Itoa(config.Server.DefaultVlan)
	handler := func(w radius.ResponseWriter, r *radius.Request) {
		username := rfc2865.UserName_GetString(r.Packet)
		username = normalizedMac(username)
		//password := rfc2865.UserPassword_GetString(r.Packet)

		vlanName := "default"
		assignedVlanId := defaultVlan

		var client Client
		var clientExists, vlanExists bool
		client, clientExists = clientMap[username]
		if clientExists {
			var vlanId string
			vlanId, vlanExists = vlanMap[client.Vlan]
			if vlanExists {
				assignedVlanId = vlanId
			}
		}

		packet := r.Response(radius.CodeAccessAccept)
		_ = rfc2868.TunnelType_Add(packet, 0, rfc3580.TunnelType_Value_VLAN)
		_ = rfc2868.TunnelMediumType_Set(packet, 0, rfc2868.TunnelMediumType_Value_IEEE802)
		_ = rfc2868.TunnelPrivateGroupID_Set(packet, 0, []byte(assignedVlanId))

		log.Printf("client=%v username=%s vlan=%s vlanName=%s clientExists=%s vlanExists=%s",
			r.RemoteAddr,
			username,
			assignedVlanId,
			vlanName,
			boolToString(clientExists),
			boolToString(vlanExists),
		)

		_ = w.Write(packet)
	}

	server := radius.PacketServer{
		Addr:         config.Server.Listen,
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(config.Server.Secret)),
	}

	log.Printf("Starting server on %s", config.Server.Listen)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func boolToString(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func normalizedMac(address string) string {
	address = strings.Replace(address, "-", "", -1)
	address = strings.Replace(address, ":", "", -1)
	address = strings.ToLower(address)
	if strings.HasPrefix(address, "0x") {
		// Remove any hex prefix
		address = address[2:]
	}
	var builder strings.Builder
	for i, r := range address {
		builder.WriteRune(r)
		if i%2 == 1 && i != len(address)-1 {
			builder.WriteRune(':')
		}
	}

	return builder.String()
}
