# Ontid ccntmract

common event format is as follows, including txhash, state, gasConsumed and notify, each native ccntmract method have different notifies.

|key|description|
|:--|:--|
|TxHash|transaction hash|
|State|1 indicates success，0 indicates fail|
|GasConsumed|gas fee consumed by this transaction|
|Notify|Notify event|

#### regIdWithPublicKey

* Usage: Register cntmid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0300000000000000000000000000000000000000", //ccntmract address of cntmid ccntmract
      "States":[
        "Register", //method name
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA" //cntmid to be registered
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

#### addKey

* Usage : Add public key to cntmid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0300000000000000000000000000000000000000", //ccntmract address of cntmid ccntmract
      "States":[
        "PublicKey", //publicKey operation
        "add", //method name
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //cntmid
        "022a06f7a4bfff93d9bbe31dfd70dbfb08263f1ea15db2ee9556688314e20e9dd7", //public key to be added
        2, //public key index
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


#### removeKey

* Usage: Remove public key

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0300000000000000000000000000000000000000", //ccntmract address of cntmid ccntmract
      "States":[
        "PublicKey", //publicKey operation
        "remove", //method name
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //cntmid
        2, //public key index
        "022a06f7a4bfff93d9bbe31dfd70dbfb08263f1ea15db2ee9556688314e20e9dd7" //public key to be removed
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

#### addRecovery

* Usage: Add recovery address

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0300000000000000000000000000000000000000", //ccntmract address of cntmid ccntmract
      "States":[
        "Recovery", //recovery operation
        "add", //method name
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //cntmid
        "d7239affb684c3c224476eb7bd52d9b2cb5e2aab" //recovery address
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

#### changeRecovery

* Usage: Change recovery address

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0300000000000000000000000000000000000000", //ccntmract address of cntmid ccntmract
      "States":[
        "Recovery", //recovery operation
        "change", //method name
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //cntmid
        "d7239affb684c3c224476eb7bd52d9b2cb5e2aab" //recovery address
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

#### regIDWithAttributes

* Usage: Register cntmid with attributes

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0300000000000000000000000000000000000000", //ccntmract address of cntmid ccntmract
      "States":[
        "Register", //method name
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA" //cntmid
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

#### addAttributes

* Usage: Add attributes

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0300000000000000000000000000000000000000", //ccntmract address of cntmid ccntmract
      "States":[
        "Attribute", //attribute operation
        "add", //method name
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //cntmid
        "", //attributes added
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

#### removeAttribute

* Usage: Delete attributes

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0300000000000000000000000000000000000000", //ccntmract address of cntmid ccntmract
      "States":[
        "Attribute", //attribute operation
        "remove", //method name
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //cntmid
        "", //attributes removed
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
