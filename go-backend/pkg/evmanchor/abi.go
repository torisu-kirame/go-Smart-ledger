package evmanchor

const ledgerAnchorABI = `[
  {
    "inputs": [
      {"internalType":"bytes32","name":"ledgerId","type":"bytes32"},
      {"internalType":"bytes32","name":"merkleRoot","type":"bytes32"},
      {"internalType":"uint64","name":"seqFrom","type":"uint64"},
      {"internalType":"uint64","name":"seqTo","type":"uint64"}
    ],
    "name": "anchor",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]`
