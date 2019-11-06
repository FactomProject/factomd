# Message Validation

The column indicates if any additional validation occurs at a given layer

| Message | P2P | Basic/Replay |  |
|----|:----------------------:|:--------------------:|----------------------|
| Ack                | [x] | [ ] / [x] | [] | 
| AddServer          | [x] | [ ] / [x] | [] | 
| Bounce             | [x] | [ ] / [x] | [] | 
| BounceReply        | [x] | [ ] / [x] | [] | 
| ChangeServerKey    | [x] | [ ] / [x] | [] | 
| CommitChain        | [x] | [ ] / [x] | [] | 
| CommitEntry        | [x] | [ ] / [x] | [] | 
| DataResponse       | [x] | [ ] / [x] | [] | 
| DBState            | [x] | [ ] / [x] | [] | 
| DBStateMissing     | [x] | [ ] / [x] | [] | 
| DBSig              | [x] | [ ] / [x] | [] | 
| EOM                | [x] | [ ] / [x] | [] | 
| FactoidTx          | [x] | [ ] / [x] | [] | 
| Heartbeat          | [x] | [ ] / [x] | [] | 
| MissingData        | [x] | [ ] / [x] | [] | 
| MissingMessage     | [x] | [ ] / [x] | [] | 
| MissingMessageResp | [x] | [ ] / [x] | [] | 
| RevealEntry        | [x] | [ ] / [x] | [] | 
| RequestBlock       | [x] | [ ] / [x] | [] | 


## Message Validation Layers

* **P2P**: Peer to peer layer   
  * Data unmarshals into a message
* **Basic**: Basic Message Validation Layer
  * Well formed
    * Data makes sense
      * Examples:
        * minute 11 is invalid
        * commit's have ec cost of reasonable amount
        * signatures match signer
        * dbstates match checkpoints (if exists)
  * Replay filter - Toss out all repeated messages    
* **??**: State required
 