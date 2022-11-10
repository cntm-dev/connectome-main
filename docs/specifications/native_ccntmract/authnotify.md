# Authentication ccntmract

common event format is as follows, including txhash, state, gasConsumed and notify, each native ccntmract method have different notifies.

|key|description|
|:--|:--|
|TxHash|transaction hash|
|State|1 indicates success，0 indicates fail|
|GasConsumed|gas fee consumed by this transaction|
|Notify|Notify event|

#### InitCcntmractAdmin

* Usage: Init admin information of a certain ccntmract through authentication ccntmract

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0600000000000000000000000000000000000000", //ccntmract address of authentication ccntmract
      "States":[
        "initCcntmractAdmin", //method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //ccntmract address of ccntmract which want to achieve authentication ccntmrol
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA" //admin cntmid if above ccntmract
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

#### Transfer

* Usage: Transfer admin to another cntmid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0600000000000000000000000000000000000000", //ccntmract address of authentication ccntmract
      "States":[
        "transfer", //method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //ccntmract address of ccntmract which want to achieve authentication ccntmrol
        true //status
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


#### AssignFuncsToRole

* Usage: Assign authentication of invoking a function in a certain ccntmract to a role

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0600000000000000000000000000000000000000", //ccntmract address of authentication ccntmract
      "States":[
        "assignFuncsToRole", //method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //ccntmract address of ccntmract which want to achieve authentication ccntmrol
        true //status
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

#### AssignOntIDsToRole

* Usage: Assign a role to a certain cntmid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0600000000000000000000000000000000000000", //ccntmract address of authentication ccntmract
      "States":[
        "assignOntIDsToRole", //method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //ccntmract address of ccntmract which want to achieve authentication ccntmrol
        true //status
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

#### Delegate

* Usage: delegate authentication to another cntmid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0600000000000000000000000000000000000000", //ccntmract address of authentication ccntmract
      "States":[
        "delegate",// method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //ccntmract address of ccntmract which want to achieve authentication ccntmrol
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from cntmid
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //to cntmid
        true //status
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

#### Withdraw

* Usage: Withdraw delegated authentication

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0600000000000000000000000000000000000000", //ccntmract address of authentication ccntmract
      "States":[
        "withdraw",// method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //ccntmract address of ccntmract which want to achieve authentication ccntmrol
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from cntmid
        "did:cntm:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //to cntmid
        true //status
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

#### VerifyToken

* Usage: Verify authentication of cntmid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "CcntmractAddress": "0600000000000000000000000000000000000000", //ccntmract address of authentication ccntmract
      "States":[
        "verifyToken", // method name
        "0700000000000000000000000000000000000000", //ccntmract address of ccntmract which want to achieve authentication ccntmrol
        "ZGlk0m9uddpBVVhDSnM3NmlqWlUzOHNlUEg5MlNuVWFvZDdQNXRVbUV4", //invoker cntmid
        "registerCandidate",// function name want to verify authentication
        true //status
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