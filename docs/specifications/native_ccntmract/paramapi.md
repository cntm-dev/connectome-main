# Native Ccntmract API : Param
* [Introduction](#introduction)
* [Ccntmract Method](#ccntmract-method)
* [How to get a parameter](#how-to-get-a-parameter)

## Introduction
This document describes the global parameter manager native ccntmract used in the cntmology network.

## Ccntmract Method

### ParamInit
Initialize the ccntmract, invoked in the genesis block.

method: init

args: nil

#### example
```
    init := states.Ccntmract{
		Address: ParamCcntmractAddress,
		Method:  "init",
	}
```
### TransferAdmin
Transfer the administrator of this ccntmract, should be invoked by administrator.

method: transferAdmin

args: smartccntmract/service/native/states.Admin

#### example
```
    var destinationAdmin states.Admin
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

args: smartccntmract/service/native/states.Admin

#### example
```
    var destinationAdmin states.Admin
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

### SetGlobalParam
Administrator set global parameter, is prepare value, won't take effect immediately.

method: setGlobalParam

args: smartccntmract/service/native/states.Params

#### example
```
    params := new(states.Params)
	*params = make(map[string]string)
	for i := 0; i < 3; i++ {
		k := "key-test" + strconv.Itoa(i) + "-" + key
		v := "value-test" + strconv.Itoa(i) + "-" + value
		(*params)[k] = v
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

### CreateSnapshot
Administrator make prepare parameter effective.

method: createSnapshot

args: nil

#### example
```
    ccntmract := &sstates.Ccntmract{
		Address: genesis.ParamCcntmractAddress,
		Method:  "createSnapshot",
	}
```

## How to get a parameter
Call the function "GetGlobalParam" to get a global parameter value.

args: smartccntmract/service/native.NativeService, the NativeServe instance<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;string, parameter name