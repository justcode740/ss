# run the cli
```
cd bls-eth-go-binary && go run examples/*.go
```

# CLI Tool Documentation

The `sscli` command-line tool is a collection of utilities for analyzing and correcting validator attestations in the Ethereum 2.0 Beacon Chain. Here is the list of available commands:

## Usage

To use `sscli`, you can either build the tool from source code or download a prebuilt binary. 

If you choose to download the prebuilt binary, simply add it to your system's `PATH` environment variable, and you can run `sscli` from anywhere in the terminal. You can find the prebuilt binary on GitHub [here](https://github.com/<user>/<repository>/releases).

## Available Commands

- `allFalse`: Fetch from Google Drive and verify all atts for all blocks, output failed verification to verificationResult/
- `args`: Print command-line arguments
- `bf`: Brute-force search pubkey for each 1-val case
- `check`: Check for all interested attestations in double / surround vote detected from beaconcha.in
- `completion`: Generate the autocompletion script for the specified shell
- `csc`: Get computed committee schedule given randaomix etc.
- `dv`: Search in-file duplicated vote from Google Drive after correction
- `fetch`: Fetch epoch for interested block under epochs/
- `fi`: Get first index of shuffled committee given randaomix etc.
- `help`: Help about any command
- `ndv`: Search duplicated vote without validator index correction, this is for detecting all potentially slashable case directly from beaconcha.in
- `recover`: Correct the wrong validator index for attestation that has false signature verification
- `search`: Brute-force search for correct signer for single validator case
- `show`: Print correction map to the console
- `shuffle`: Shuffle duplicate votes
- `sr`: Compute and print signing root for a specific attestation
- `ss`: Use unvoted validators as a restrained space for faster search pubkey for correcting signing validator sets
- `unvote`: Find all unvoted validators in an epoch defined by input block
- `vc`: Verify all correction made from computed committee schedule by checking the signature
- `vdv`: Search all verifiable in-file double vote after correction from Google Drive, expect all output should be slashed otherwise staking-slashing model has trouble
- `verify`: Verify attestations that identified by validatorIdx and block number, print result to console
- `verifyAll`: Verify all attestations in a specified block, if the block attestation data doesn't exist in local, automatically fetch to test/
- `verifyDrive`: Fetch from Google Drive and verify all atts for all blocks, output failed verification to verificationResult/

## Flags

- `-h`, `--help`: Show help message for `sscli`



