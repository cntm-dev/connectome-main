# Ont ccntmract

common event format is as follows, including txhash, state, gasConsumed and notify, each native ccntmract method have different notifies.

|key|description|
|:--|:--|
|TxHash|transaction hash|
|State|1 indicates success，0 indicates fail|
|GasConsumed|gas fee consumed by this transaction|
|Notify|Notify event|

#### Transfer

* Usage: Transfer cntm

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
        "transfer",// method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from address
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //to address
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

#### TransferFrom

* Usage: Transfer from cntm

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
        "transfer",// method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from address
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //to address
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
