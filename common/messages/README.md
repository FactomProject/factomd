# Message Validation

The column indicates if any additional validation occurs at a given layer

| Message | P2P | Basic |  |
|----|:----------------------:|:--------------------:|----------------------|
| Ack                | [x] | [] | [] | 
| AddServer          | [x] | [] | [] | 
| Bounce             | [x] | [] | [] | 
| BounceReply        | [x] | [] | [] | 
| ChangeServerKey    | [x] | [] | [] | 
| CommitChain        | [x] | [] | [] | 
| CommitEntry        | [x] | [] | [] | 
| DataResponse       | [x] | [] | [] | 
| DBState            | [x] | [] | [] | 
| DBStateMissing     | [x] | [] | [] | 
| DBSig              | [x] | [] | [] | 
| EOM                | [x] | [] | [] | 
| FactoidTx          | [x] | [] | [] | 
| Heartbeat          | [x] | [] | [] | 
| MissingData        | [x] | [] | [] | 
| MissingMessage     | [x] | [] | [] | 
| MissingMessageResp | [x] | [] | [] | 
| RevealEntry        | [x] | [] | [] | 
| RequestBlock       | [x] | [] | [] | 


## Message Validation Layers

* **P2P**: Peer to peer layer   
  * Data unmarshals into a message
* **Basic**: Basic Message Validation Layer
  * Well formed
    * Data makes sense
      * minute 11 is invalid
      * commit's have ec cost of reasonable amount
      * signatures match signer
      * dbstates match checkpoints (if exists)
  * Replay filter - Toss out all repeated messages    
* **??**: State required
 