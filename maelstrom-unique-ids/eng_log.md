# Debugging duplicate initialization acknowledgements 

## Observations
- used a init handler and initialize generator inside to user the node id in the generated ids
- Only one node was processing most of the request (99)
- other nodes were idle 
- timeouts were seen after evaluation was completed with `availability` metric failed 


## Root cause
- in the `init` handler multiple reply were sent , first auto `init_ok` by maelstrom itself and second custom reply by me in the handler
- this caused the failure on the maelstrom process 
- maelstrom creates client entities which are removed after receiving first initializtion acknowledge but for second it fails it is received the client entity is removed 
- this crashed that initialized client and the nodes were disconnected 
- but this based on timing if second init was received before client destroys the entity then it would work fine 
- but sometimes all nodes would fails and which causes the behaviour observed