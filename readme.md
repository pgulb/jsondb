# github.com/pgulb/jsondb  
  
Embedded in-memory disk-persisted JSON-based Go string database.  
  
Example usage in example/example.go  
Communication with jsondb is done by running looping function in goroutine  
and sending instructions through channels.  
  
key-value pairs are divided into families (like databases in Redis)  
  
Avalaible commands:  
 - set - sets key to value in specific keyFamily  
 - get - gets specific key from keyFamily  
 - list - lists keyFamilys  
 - listkeys - lists keys in specific keyFamily  
 - quit - close loop  
  
Example request sent into channel:  
```go
structures.Request{
	KeyFamily: "keyFamilyName",
	Key:       "keyName",
	Value:     "qwe123",
	Action:    "set",
}
```
  