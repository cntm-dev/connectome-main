(module
 (type $FUNCSIG$i (func (result i32)))
 (type $FUNCSIG$iiiii (func (param i32 i32 i32 i32) (result i32)))
 (import "env" "callCcntmract" (func $callCcntmract (param i32 i32 i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 16) "add\00")
 (data (i32.const 32) "sum\00")
 (export "memory" (memory $0))
 (export "testCall" (func $testCall))
 (func $testCall (; 1 ;) (param $0 i32) (param $1 i32) (result i32)
  (call $callCcntmract
   (i32.const 16)
   (i32.const 32)
   (get_local $0)
   (get_local $1)
  )
 )
)
