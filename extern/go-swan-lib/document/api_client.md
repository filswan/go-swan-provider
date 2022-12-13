# Groups
* [Ipfs](#Ipfs)
  * [IpfsUploadCarFile](#IpfsUploadCarFile)
* [Lotus](#Lotus)
  * [LotusGetClient](#LotusGetClient)
  * [LotusClientCalcCommP](#LotusClientCalcCommP)
  * [LotusClientImport](#LotusClientImport)
  * [LotusClientGenCar](#LotusClientGenCar)
  * [LotusGetMinerConfig()](#LotusGetMinerConfigs())
  * [LotusProposeOfflineDeal](#LotusProposeOfflineDeal)
* [ExecOsCmd](#ExecOsCmd)
  * [ExecOsCmd2Screen](#ExecOsCmd2Screen)
  * [ExecOsCmd](#ExecOsCmd)
  * [ExecOsCmdBase](#ExecOsCmdBase)
* [Http](#[Http)
  * [HttpPostNoToken](#HttpPostNoToken)
  * [HttpPost](#HttpPost)
  * [HttpGetNoToken](#HttpGetNoToken)
  * [HttpGet](#HttpGets)
  * [HttpPut](#HttpPut)
  * [HttpDelete](#HttpDelete)
  * [httpRequest](#httpRequest)
  * [HttpPutFile](#HttpPutFile)
  * [HttpPostFile](#HttpPostFile)
  * [HttpRequestFile](#HttpRequestFile)
* [Swan](#Swan)
  * [SwanGetJwtToken](#SwanGetJwtToken)
  * [SwanGetClient](#SwanGetClient)
  * [SwanGetOfflineDeals](#SwanGetOfflineDeal)
  * [SwanUpdateOfflineDealStatus](#SwanUpdateOfflineDealStatus)
  * [SwanCreateTask](#SwanCreateTask)
  * [SwanGetTasks](#SwanGetTasks)
  * [SwanGetAssignedTasks](#SwanGetAssignedTasks)
  * [SwanGetOfflineDealsByTaskUuid](#SwanGetOfflineDealsByTaskUuid)
  * [SwanUpdateTaskByUuid](#SwanUpdateTaskByUuid)
  * [SwanUpdateAssignedTask](#SwanUpdateAssignedTask)


## Ipfs
### IpfsUploadCarFile

Definition:
```shell
func IpfsUploadCarFile(carFilePath string) (*string, error)
```

Outputs:
```shell
*string: car file hash
error: error or nil
```

## Lotus
### LotusGetClients

Definition:
```shell
func LotusGetClient(apiUrl, accessToken string) (*LotusClient, error)
apiUrl  string   #lotus node api url, such as http://[ip]:[port]/rpc/v0
accessToken  string  #lotus node access token, should have admin privilege
```

Outputs:
```shell
*LotusClient #structure including access info for lotus node
error: error or nil
```

### LotusClientCalcCommP

Definition:
```shell
func (lotusClient *LotusClient) LotusClientCalcCommP(filepath string) *string
```

Outputs:
```shell
*string  #piece cid, or nil when cannot get the info required
```

### LotusClientImport

Definition:
```shell
func (lotusClient *LotusClient) LotusClientImport(filepath string, isCar bool) (*string, error)
```

Outputs:
```shell
*string  #piece cid, or nil when cannot get the info required
```

### LotusClientGenCar

Definition:
```shell
func (lotusClient *LotusClient) LotusClientGenCar(srcFilePath, destCarFilePath string, srcFilePathIsCar bool) error
```

Outputs:
```shell
error  #error or nils
```

### LotusGetMinerConfig

Definition:
```shell
func LotusGetMinerConfig(minerFid string) (*decimal.Decimal, *decimal.Decimal, *string, *string)
```

Outputs:
```shell
*decimal.Decimal  # price
*decimal.Decimal  # verified price
*string  # max piece size
*string  # min piece size
```

### LotusProposeOfflineDeal

Definition:
```shell
func LotusProposeOfflineDeal(carFile model.FileDesc, cost decimal.Decimal, pieceSize int64, dealConfig model.ConfDeal, relativeEpoch int) (*string, *int, error)
```

Outputs:
```shell
*string  # deal cid
*int  # start epoch
error # error or nil
```

## ExecOsCmd
### ExecOsCmd2Screen

Definition:
```shell
func ExecOsCmd2Screen(cmdStr string, checkStdErr bool) (string, error)
```

Outputs:
```shell
string  # standard output
error # error or nil
```

### ExecOsCmd

Definition:
```shell
func ExecOsCmd(cmdStr string, checkStdErr bool) (string, error)
```

Outputs:
```shell
string  # standard output
error # error or nil
```


### ExecOsCmdBase

Definition:
```shell
func ExecOsCmdBase(cmdStr string, out2Screen bool, checkStdErr bool) (string, error)
```

Outputs:
```shell
string  # standard output
error # error or nil
```

## Http
### HttpPostNoToken

Definition:
```shell
func HttpPostNoToken(uri string, params interface{}) string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpPost

Definition:
```shell
func HttpPost(uri, tokenString string, params interface{}) string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpGetNoToken

Definition:
```shell
func HttpGetNoToken(uri string, params interface{}) string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpGet

Definition:
```shell
func HttpGet(uri, tokenString string, params interface{}) string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpPut

Definition:
```shell
func HttpPut(uri, tokenString string, params interface{}) string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpDelete

Definition:
```shell
func HttpDelete(uri, tokenString string, params interface{}) string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### httpRequest

Definition:
```shell
func httpRequest(httpMethod, uri, tokenString string, params interface{}) string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpPutFile

Definition:
```shell
func HttpPutFile(url string, tokenString string, paramTexts map[string]string, paramFilename, paramFilepath string) (string, error)
```

Definition:
```shell
string  # result from web api request, if error, then ""
error # error or nil
```

### HttpPostFile

Definition:
```shell
func HttpPostFile(url string, tokenString string, paramTexts map[string]string, paramFilename, paramFilepath string) (string, error)
```

Outputs:
```shell
string  # result from web api request, if error, then ""
error # error or nil
```

### HttpRequestFile

Definition:
```shell
func HttpRequestFile(httpMethod, url string, tokenString string, paramTexts map[string]string, paramFilename, paramFilepath string) (string, error)
```

Outputs:
```shell
string  # result from web api request, if error, then ""
error # error or nil
```

## Swan
### SwanGetJwtToken

Definition:
```shell
func (swanClient *SwanClient) SwanGetJwtToken(apiKey, accessToken string) error
```

Outputs:
```shell
string  # result from web api request, if error, then ""
error # error or nil
```

### SwanGetClient

Definition:
```shell
func SwanGetClient(apiUrl, apiKey, accessToken string) (*SwanClient, error)
```

Outputs:
```shell
*SwanClient
error
```

### SwanGetOfflineDeals

Definition:
```shell
func (swanClient *SwanClient) SwanGetOfflineDeals(minerFid, status string, limit ...string) []model.OfflineDeal
```

Outputs:
```shell
[]model.OfflineDeal
```

### SwanUpdateOfflineDealStatus

Definition:
```shell
func (swanClient *SwanClient) SwanUpdateOfflineDealStatus(dealId int, status string, statusInfo ...string) bool
```

Outputs:
```shell
bool
```

### SwanCreateTask

Definition:
```shell
func (swanClient *SwanClient) SwanCreateTask(task model.Task, csvFilePath string) (*SwanCreateTaskResponse, error)
```

Outputs:
```shell
*SwanCreateTaskResponse
error
```

### SwanGetTasks

Definition:
```shell
func (swanClient *SwanClient) SwanGetTasks(limit *int) (*GetTaskResult, error)
```

Outputs:
```shell
*GetTaskResult
error
```

### SwanGetAssignedTasks

Definition:
```shell
func (swanClient *SwanClient) SwanGetAssignedTasks() ([]model.Task, error)
```

Outputs:
```shell
[]model.Task
error
```

### SwanGetOfflineDealsByTaskUuid

Definition:
```shell
func (swanClient *SwanClient) SwanGetOfflineDealsByTaskUuid(taskUuid string) (*GetOfflineDealsByTaskUuidResult, error)
```

Outputs:
```shell
*GetOfflineDealsByTaskUuidResult
error
```

### SwanUpdateTaskByUuid

Definition:
```shell
func (swanClient *SwanClient) SwanUpdateTaskByUuid(taskUuid string, minerFid string, csvFilePath string) error
```

Outputs:
```shell
*GetOfflineDealsByTaskUuidResult
error
```

### SwanUpdateAssignedTask

Definition:
```shell
func (swanClient *SwanClient) SwanUpdateAssignedTask(taskUuid, status, csvFilePath string) (*SwanCreateTaskResponse, error)
```

Outputs:
```shell
*SwanCreateTaskResponse
error
```
