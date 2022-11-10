# Governance ccntmract

common event format is as follows, including txhash, state, gasConsumed and notify, each native ccntmract method have different notifies.

|key|description|
|:--|:--|
|TxHash|transaction hash|
|State|1 indicates success，0 indicates fail|
|GasConsumed|gas fee consumed by this transaction|
|Notify|Notify event|

#### RegisterCandidate

* Usage: Register to become a candidate node

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the auth ccntmract
    {
      "CcntmractAddress": "0600000000000000000000000000000000000000", //ccntmract address of auth ccntmract
      "States":[
        "verifyToken", //method name
        "0700000000000000000000000000000000000000", //governance ccntmract address
        "ZGlk0m9uddpBVVhDSnM3NmlqWlUzOHNlUEg5MlNuVWFvZDdQNXRVbUV4", //invoker cntmid
        "registerCandidate",// authorize function name
        true //status
      ]
    },
    //notify of cntm transfer
    {
      "CcntmractAddress": "0100000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer",// method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from address
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //to address
        100 //transfer amount
      ]
    },
    //notify of cntm transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer",// method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from address
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //to address
        100 //transfer amount
      ]
    },
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### UnRegisterCandidate

* Usage: Cancel register candidate request

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### QuitNode

* Usage: Quit candidate node

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### AuthorizeForPeer

* Usage: Authorize cntm to a candidate node

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of cntm transfer
    {
      "CcntmractAddress":"0100000000000000000000000000000000000000", //cntm ccntmract address
      "State":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from address
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //to address
        10000000 //transfer amount
      ]
    },
    //unbounded cntm transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", // method name
        "AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV", //cntm ccntmract address
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker address
        10000000 //unbounded cntm amount
      ]
    },
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### UnAuthorizeForPeer

* Usage: Cancel the authorize cntm to a candidate node

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### Withdraw

* Usage: Withdraw deposit cntm

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of cntm transfer
    {
      "CcntmractAddress": "0100000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", // method name
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK",// governance ccntmract
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA",// invoker address
        10000000 // withdraw amount
      ]
    },
    //unbounded cntm transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", // method name
        "AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV", //cntm ccntmract address
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker address
        10000000 //unbounded cntm amount
      ]
    },
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### WithdrawOng

* Usage: Withdraw unbounded cntm

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of cntm transfer to trigger unbounded cntm
    {
      "CcntmractAddress": "0100000000000000000000000000000000000000",
      "States":[
        "transfer", //method name
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address
        1 //fixed amount
      ]
    },
    //unbounded cntm transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", // method name
        "AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV", //cntm ccntmract address
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker address
        10000000 //unbounded cntm amount
      ]
    },
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### AddInitPos

* Usage: Add node's init pos

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of cntm transfer
    {
      "CcntmractAddress": "0100000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker address
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance address
        1000 //add init pos amount
      ]
    },
    //unbounded cntm transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", // method name
        "AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV", //cntm ccntmract address
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker address
        10000000 //unbounded cntm amount
      ]
    },
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### ReduceInitPos

* Usage: Reduce node's init pos

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of gas fee transfer
    {
      "CcntmractAddress": "0200000000000000000000000000000000000000", //cntm ccntmract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance ccntmract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```
