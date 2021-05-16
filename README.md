# QUIC Adapter
This is an Abstract Symbols <-> Concrete Symbol adapter for the QUIC protocol based on quic-tracker.

### Interesting Components:
* adapter/adapter.go -> The main interface for the learner, start point for requests.
* adapter/abstract.go -> Implementation of abstract alphabet.
* adapter/concrete.go -> Implementation of concrete alphabet.
* agents/ -> Collection of agents responsible for each aspect of the protocol.
* connection.go -> Main protocol state.
