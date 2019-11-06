# Basic Message Validation

Basic Message validation is a module that sits right off the P2P network. It handles all inbound messages, and performs a `WellFormed` check + replay filter check. The `WellFormed` check is all validation that can happen without any state. The replay filter will reject any message that has already been seen by the BMV.