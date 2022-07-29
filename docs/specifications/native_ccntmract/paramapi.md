# Native Ccntmract API : Param
* [Introduction](#introduction)
* [Ccntmract Method](#ccntmract-method)

## Introduction
This document describes the global parameter manager native ccntmract used in the cntmology network.

## Ccntmract Method

### ParamInit
Initialize the ccntmract, invoked in the genesis block.

method: init

args: nil

return: bool

#### example
```
    init := sstates.Ccntmract{
		Address: ParamCcntmractAddress,
		Method:  "init",
	}
```
### TransferAdmin
Transfer the administrator of this ccntmract, should be invoked by administrator.

method: transferAdmin

args: smartccntmract/service/native/global_params.Admin

return: bool

#### example
```
    var destinationAdmin global_params.Admin
	address, _ := common.AddressFromBase58("TA4knXiWFZ8K4W3e5fAnoNntdc5G3qMT7C")
	copy(destinationAdmin[:], address[:])
	adminBuffer := new(bytes.Buffer)
	if err := destinationAdmin.Serialize(adminBuffer); err != nil {
		fmt.Println("Serialize admins struct error.")
		os.Exit(1)
	}
	ccntmract := &sstates.Ccntmract{
		Address: genesis.ParamCcntmractAddress,
		Method:  "transferAdmin",
		Args:    adminBuffer.Bytes(),
	}
```

### AcceptAdmin
Accept administrator permission of the ccntmract.

method: acceptAdmin

args: smartccntmract/service/native/global_params.Admin

return: bool

#### example
```
    var destinationAdmin global_params.Admin
	address, _ := common.AddressFromBase58("TA4knXiWFZ8K4W3e5fAnoNntdc5G3qMT7C")
	copy(destinationAdmin[:], address[:])
	adminBuffer := new(bytes.Buffer)
	if err := destinationAdmin.Serialize(adminBuffer); err != nil {
		fmt.Println("Serialize admin struct error.")
		os.Exit(1)
	}

	ccntmract := &sstates.Ccntmract{
		Address: genesis.ParamCcntmractAddress,
		Method:  "acceptAdmin",
		Args:    adminBuffer.Bytes(),
	}
```
### SetOperator
Administrator set operator of the ccntmract.

method: setOperator

args: smartccntmract/service/native/global_params.Admin

return: bool
#### example
```
    var destinationOperator global_params.Admin
	address, _ := common.AddressFromBase58("TA4knXiWFZ8K4W3e5fAnoNntdc5G3qMT7C")
	copy(destinationOperator[:], address[:])
	adminBuffer := new(bytes.Buffer)
	if err := destinationOperator.Serialize(adminBuffer); err != nil {
		fmt.Println("Serialize admin struct error.")
		os.Exit(1)
	}

	ccntmract := &sstates.Ccntmract{
		Address: genesis.ParamCcntmractAddress,
		Method:  "acceptAdmin",
		Args:    adminBuffer.Bytes(),
	}
```

### SetGlobalParam
Operator set global parameter, is prepare value, won't take effect immediately.

method: setGlobalParam

args: smartccntmract/service/native/global_params.Params

return: bool

#### example
```
    params := new(global_params.Params)
	*params = make(map[string]string)
	for i := 0; i < 3; i++ {
		k := "key-test" + strconv.Itoa(i) + "-" + key
		v := "value-test" + strconv.Itoa(i) + "-" + value
		(*params) = append(*params, &global_params.Param{k,v})
	}
	paramsBuffer := new(bytes.Buffer)
	if err := params.Serialize(paramsBuffer); err != nil {
		fmt.Println("Serialize params struct error.")
		os.Exit(1)
	}

	ccntmract := &sstates.Ccntmract{
		Address: genesis.ParamCcntmractAddress,
		Method:  "setGlobalParam",
		Args:    paramsBuffer.Bytes(),
	}
```

### GetGlobalParam
Get global parameter, the method will return smartccntmract/service/native/global_params.Params

method: getGlobalParam

args: smartccntmract/service/native/global_params.ParamNameList

return: array

#### example
```
    nameList := new(global_params.ParamNameList)
	for i := 0; i < 3; i++ {
		k := "key-test" + strconv.Itoa(i) + "-" + key
		(*nameList) = append(*nameList, k)
	}
	nameListBuffer := new(bytes.Buffer)
	if err := nameList.Serialize(nameListBuffer); err != nil {
		fmt.Println("Serialize ParamNameList struct error.")
		os.Exit(1)
	}
	ccntmract := &sstates.Ccntmract{
		Address: genesis.ParamCcntmractAddress,
		Method:  "getGlobalParam",
		Args:    nameListBuffer.Bytes(),
	}
```

### CreateSnapshot
Operator make prepare parameter effective.

method: createSnapshot

args: nil

return: bool

#### example
```
    ccntmract := &sstates.Ccntmract{
		Address: genesis.ParamCcntmractAddress,
		Method:  "createSnapshot",
	}
```
