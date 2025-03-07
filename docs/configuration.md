# Configuration
Vouch can be configured through environment, command-line or configuration file.  In the case of conflicting configuration the order of precedence is:

  - command-line; then
  - environment; then
  - configuration file.

# The configuration file
Vouch's configuration file can be written in JSON or YAML.  The file can either be in the user's home directory, in which case it will be called `.vouch.json` (or `.vouch.yml`), or it can be in a directory specified by the command line option `--base-dir` or environment variable `VOUCH_BASE_DIR`, in which case it will be called `vouch.json` (or `vouch.yml`).

A sample configuration file in YAML with is shown below:

```
# log-file is the location for Vouch log output.  If this is not provided logs will be written to the console.
log-file: /home/me/vouch.log
# log-level is the global log level for Vouch logging.
# Overrides can be set at any sub-level, giving fine-grained control over the specific
# information logged.
log-level: Debug

# beacon-node-address is the address of the beacon node.  Can be lighthouse, nimbus, prysm or teku.
# Overridden by beacon-node-addresses if present.
beacon-node-address: localhost:4000

# beacon-node-addresseses is the list of address of the beacon nodes.  Can be lighthouse, nimbus, prysm or teku.
# If multiple addresses are supplied here it makes Vouch resilient in the situation where a beacon
# node goes offline entirely.  If this occurs to the currently used node then the next in the list will
# be used.  If a beacon node comes back online it is added to the end of the list of potential nodes to
# use.
#
# Note that some beacon nodes have slightly different behavior in their events.  As such, users should
# ensure they are happy with the event output of all beacon nodes in this list.
beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052 ]

# metrics is the module that logs metrics, in this case using prometheus.
metrics:
  prometheus:
    # log-level is the log level for this module, over-riding the global level.
    log-level: warn
    # listen-address is the address on which prometheus listens for metrics requests.
    listen-address: 0.0.0.0:8081

# graffiti provides graffiti data.  Full details are in the separate document.
graffiti:
  static:
    value: My graffiti

# scheduler handles the scheduling of Vouch's operations.
scheduler:
  # style can be 'basic' (deprecated) or 'advanced' (default).  Do not use the basic scheduler unless instructed.
  style: advanced

# submitter submits data to beacon nodes.  If not present the nodes in beacon-node-address above will be used.
submitter:
  # style can currently only be 'multinode'
  style: multinode
  aggregateattestation:
    # beacon-node-addresses are the addresses to which to submit aggregate attestations.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  attestation:
    # beacon-node-addresses are the addresses to which to submit attestations.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  beaconblock:
    # beacon-node-addresses are the addresses to which to submit beacon blocks.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  beaconcommitteesubscription:
    # beacon-node-addresses are the addresses to which to submit beacon committee subscriptions.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  proposalpreparation:
    # beacon-node-addresses are the addresses to which to submit beacon proposal preparations.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  synccommitteecontribution:
    # beacon-node-addresses are the addresses to which to submit beacon sync committee contributions.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  synccommitteemessage:
    # beacon-node-addresses are the addresses to which to submit beacon sync committee messages.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  synccommitteesubscription:
    # beacon-node-addresses are the addresses to which to submit beacon sync committee subscriptions.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]

# fee recipient provides information about the fee recipient for block proposals.  Advanced configuration
# information is available in the documentation.
feerecipient:
  default-address: '0x0000000000000000000000000000000000000001'

# strategies provide advanced strategies for dealing with multiple beacon nodes
strategies:
  # The beaconblockproposal strategy obtains beacon block proposals from multiple sources.
  beaconblockproposal:
    # style can be 'best', which obtains blocks from all nodes and selects the best, or 'first', which uses the first returned
    style: best
    # beacon-node-addresses are the addresses from which to receive beacon block proposals.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
    # timeout defines the maximum amount of time the strategy will wait for a response.  As soon as a response from all beacon
    # nodes has been obtained,the strategy will return with the best.  Half-way through the timeout period, Vouch will check to see
    # if there have been any responses from the beacon nodes, and if so will return with the best.
    # This allows Vouch to remain responsive in the situation where some beacon nodes are significantly slower than others, for
    # example if one is remote.
    timeout: 2s
  # The attestationdata strategy obtains attestation data from multiple sources.
  attestationdata:
    # style can be 'best', which obtains attestation data from all nodes and selects the best, or 'first', which uses the first returned
    style: best
    # beacon-node-addresses are the addresses from which to receive attestation data.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  # The aggregateattestation strategy obtains aggregate attestations from multiple sources.
  # Note that the list of nodes here must be a subset of those in the attestationdata strategy.  If not, the nodes will not have
  # been gathering the attestations to aggregate and will error when the aggregate request is made.
  aggregateattestation:
    # style can be 'best', which obtains aggregates from all nodes and selects the best, or 'first', which uses the first returned
    style: best
    # beacon-node-addresses are the addresses from which to receive aggregate attestations.
    # Note that prysm nodes are not supported at current in this strategy.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
  # The synccommitteecontribution strategy obtains sync committee contributions from multiple sources.
  synccommitteecontribution:
    # style can be 'best', which obtains contributions from all nodes and selects the best, or 'first', which uses the first returned
    style: best
    # beacon-node-addresses are the addresses from which to receive sync committee contributions.
    beacon-node-addresses: [ localhost:4000, localhost:5051, localhost:5052]
```

## Hierarchical configuration.
A number of items in the configuration are hierarchical.  If not stated explicitly at a point in the configuration file, Vouch will move up the levels of configuration to attempt to find the relevant information.  For example, when searching for the value `submitter.attestation.multinode.beacon-node-addresses` the following points in the configuration will be checked:

  - `submitter.attestation.multinode.beacon-node-addresses`
  - `submitter.attestation.beacon-node-addresses`
  - `submitter.beacon-node-addresses`
  - `beacon-node-addresses`

Vouch will use the first value obtained.  Continuing the example, if a configuration file is set up as follows:

```
beacon-node-addresses: [ localhost:4000, localhost:5051 ]
strategies:
  beacon-node-address: [ localhost: 5051 ]
  beaconblockproposal:
    style: best
    beacon-node-addresses: [ localhost:4000 ]
submitter:
  style: multinode
  beaconblock:
    multinode:
      beacon-node-addresses: [ localhost:4000, localhost:9000 ]
```

Then the configuration will resolve as follows:
  - `beacon-node-addresses` resolves to `[ localhost:4000, localhost:5051 ]` with a direct match
  - `strategies.attestationdata.best.beacon-node-addresses` resolves `[ localhost:5051 ]` at `strategies.beacon-node-addresses`
  - `strategies.beaconblockproposal.best.beacon-node-addresses` resolves `[ localhost:4000 ]` at `strategies.beacon-node-addresses`
  - `submitter.beaconblock.multinode.beacon-node-addresses` resolves `[ localhost:4000, localhost:9000 ]` with a direct match
  - `submitter.attestation.multinode.beacon-node-addresses` resolves `[ localhost:4000, localhost:5051 ]` at `beacon-node-addresses`

Hierarchical configuration provides a simple way of setting defaults and overrides, and is available for `beacon-node-addresses`, `log-level` and `process-concurrency` configuration values.

## Logging
Vouch has a modular logging system that allows different modules to log at different levels.  The available log levels are:

  - **Fatal**: messages that result in Vouch stopping immediately;
  - **Error**: messages due to Vouch being unable to fulfil a valid process;
  - **Warning**: messages that result in Vouch not completing a process due to transient or user issues;
  - **Information**: messages that are part of Vouch's normal startup and shutdown process;
  - **Debug**: messages when one of Vouch's processes diverge from normal operations;
  - **Trace**: messages that detail the flow of Vouch's normal operations; or
  - **None**: no messages are written.

### Global level
The global level is used for all modules that do not have an explicit log level.  This can be configured using the command line option `--log-level`, the environment variable `VOUCH_LOG_LEVEL` or the configuration option `log-level`.

### Module levels
Modules levels are used for each module, overriding the global log level.  The available modules are:

  - **accountmanager** access to validating accounts
  - **attestationaggregator** aggregating attestations
  - **attester** attesting to blocks
  - **beaconcommitteesubscriber** subscribing to beacon committees
  - **beaconblockproposer** proposing beacon blocks
  - **chaintime** calculations for time on the blockchain (start of slot, first slot in an epoch _etc._)
  - **controller** control of which jobs occur when
  - **graffiti** provision of graffiti for proposed blocks
  - **majordomo** accesss to secrets
  - **scheduler** starting internal jobs such as proposing a block at the appropriate time
  - **signer** carries out signing activities
  - **strategies.beaconblockproposer** decisions on how to obtain information from multiple beacon nodes
  - **strategies.synccommitteecontribution** decisions on how to obtain information from multiple beacon nodes
  - **submitter** decisions on how to submit information to multiple beacon nodes
  - **validatorsmanager** obtaining validator state from beacon nodes and providing it to other modules

This can be configured using the environment variables `VOUCH_<MODULE>_LOG_LEVEL` or the configuration option `<module>.log-level`.  For example, the controller module logging could be configured using the environment variable `VOUCH_CONTROLLER_LOG_LEVEL` or the configuration option `controller.log-level`.

## Advanced options
Advanced options can change the performance of Vouch to be severely detrimental to its operation.  It is strongly recommended that these options are not changed unless the user understands completely what they do and their possible performance impact.

### controller.max-attestation-delay
This is a duration parameter, that defaults to `4s`.  It defines the maximum time that Vouch will wait from the start of a slot for a block before attesting on the basis that the slot is empty.

### controller.attestation-aggregation-delay
This is a duration parameter, that defaults to `8s`.  It defines the time that Vouch will wait from the start of a slot before aggregating existing attestations.

### controller.max-sync-committee-message-delay
This is a duration parameter, that defaults to `4s`.  It defines the maximum time that Vouch will wait from the start of a slot for a block before generating sync committee messages on the basis that the slot is empty.

### controller.sync-committee-aggregation-delay
This is a duration parameter, that defaults to `8s`.  It defines the time that Vouch will wait from the start of a slot before aggregating existing sync committee messages.
