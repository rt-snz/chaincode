package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}


func (t *SimpleChaincode) create_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var initial_asset int;
	var user,password,key_password,key_balance string;
	var err error;
	
	user = args[0];
	password = args[1];
	initial_asset = 10000;	

	key_password = user + "_p";
	key_balance = user + "_b";
	
	err = stub.PutState(key_password, []byte(password))

	if err != nil {
		return nil, err
	}

	err = stub.PutState(key_balance, []byte(strconv.Itoa(initial_asset)))
	
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var sender, reciever string          // Entities
	var senderAmount, recieverAmount int // Asset holdings
	var X int                            // Transaction value
	var err error

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	sender = args[0]
	reciever = args[1]

	isEnableSender := IsExistUser(sender)
	isEnableReciever := IsExistUser(reciever)

	if isEnableSender == false || isEnableReciever == false {
		return nil, errors.New("No exsit user")
	}

	senderAmountbytes, err := stub.GetState(sender)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	if senderAmountbytes == nil {
		senderAmount = 0
	} else {
		senderAmount, _ = strconv.Atoi(string(senderAmountbytes))
	}

	recieverAmountbytes, err := stub.GetState(reciever)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	if recieverAmountbytes == nil {
		recieverAmount = 0
	} else {
		recieverAmount, _ = strconv.Atoi(string(recieverAmountbytes))
	}

	X, err = strconv.Atoi(args[2])
	senderAmount = senderAmount - X
	recieverAmount = recieverAmount + X
	fmt.Printf("senderAmount = %d, recieverAmount = %d\n", senderAmount, recieverAmount)

	err = stub.PutState(sender, []byte(strconv.Itoa(senderAmount)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState(reciever, []byte(strconv.Itoa(recieverAmount)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Running delete")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	A := args[0]

	err := stub.DelState(A)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil
}


func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Invoke called, determining function")

	if function == "transfer" {
		return t.transfer(stub, args)
	} else if function == "create_user" {
		return t.create_user(stub, args)	
	} else if function == "init" {
		return t.Init(stub, function, args)
	} else if function == "delete" {
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Run called, passing through to Invoke (same function)")

	if function == "init" {
		return t.Init(stub, function, args)
	} else if function == "delete" {
		return t.delete(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return nil, errors.New(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return Avalbytes, nil

}

func (t *SimpleChaincode) cert(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var user, password, key_password string
	var err error;

	user = args[0];
	key_password = user + "_p";
	password = args[1];
	
	password_state_bytes, err := stub.GetState(key_password)

	if err != nil {
		jsonResp := "{\"Error\":\"user is not registered" + "\"}"
		return nil, errors.New(jsonResp)
	}
	
	if string(password_state_bytes) == password {
		return []byte("200"),nil
	}else{
		jsonResp := "{\"Error\":\"password does not match" + password_state_bytes + " " + password + "\"}"
		return nil, errors.New(jsonResp)	
	}
		return nil, nil
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	fmt.Printf("Query called, determining function")

	if function == "getBalance" {
		return t.getBalance(stub, args)
	} else if function == "cert" {
		return t.cert(stub, args)
	}

	jsonResp := "{\"Error\":\"Received unknown function Query" + "\"}"
	return nil, errors.New(jsonResp)
}

func IsExistUser(userId string) bool {
	var client, _ = NewAPIClient("https://b23476f36d234c06aff5e3f1822e3c03-vp0.us.blockchain.ibm.com:5003/", "", "", nil)
	registrarURL := "registrar/" + userId
	var request, _ = client.NewRequest("GET", registrarURL, nil)
	var response, responseError = client.HTTPClient.Do(request)
	if responseError != nil {
		return false
	}
	fmt.Println("response")
	responseByteArray, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	fmt.Println("unmarshal")
	var registrarResult interface{}
	unmarshalError := json.Unmarshal(responseByteArray, &registrarResult)
	if unmarshalError != nil {
		return false
	}
	fmt.Printf("%+v\n", registrarResult)
	convertResult := registrarResult.(map[string]interface{})["OK"]
	var isExist bool = false
	if convertResult != nil {
		resultCode := convertResult.(string)
		fmt.Println(resultCode)
		isExist = true
	} else {
		fmt.Println("login error")
	}

	return isExist
}

type APIClient struct {
	URL        *url.URL
	HTTPClient *http.Client

	Username, Password string
	Logger             *log.Logger
}

func NewAPIClient(urlString, username, password string, logger *log.Logger) (*APIClient, error) {
	fmt.Println("start Creating NewAPIClient")
	parserdURL, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, errors.New("faild to parse URL: " + urlString)
	}
	fmt.Println("URL OK.")

	var discardLogger = log.New(ioutil.Discard, "", log.LstdFlags)
	if logger == nil {
		logger = discardLogger
	}
	fmt.Println("Logger OK.")
	apiClient := APIClient{}
	apiClient.Username = username
	apiClient.Password = password
	apiClient.URL = parserdURL
	apiClient.Logger = logger
	apiClient.HTTPClient = &http.Client{Timeout: time.Duration(10) * time.Second}
	return &apiClient, nil
}

func (apiClient *APIClient) NewRequest(method, spath string, body io.Reader) (*http.Request, error) {
	u := *apiClient.URL
	u.Path = path.Join(apiClient.URL.Path, spath)

	request, error := http.NewRequest(method, u.String(), body)
	if error != nil {
		return nil, error
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	return request, nil
}

func DecodeBody(response *http.Response, out interface{}) error {
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	return decoder.Decode(out)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}