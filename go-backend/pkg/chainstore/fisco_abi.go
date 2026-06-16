package chainstore

// ledgerRegistryABI — putState/getState only (KV mirror of MiniLedger world_state).
const ledgerRegistryABI = `[
  {"inputs":[{"internalType":"string","name":"ledgerId","type":"string"},{"internalType":"string","name":"key","type":"string"},{"internalType":"bytes","name":"value","type":"bytes"}],"name":"putState","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"string","name":"ledgerId","type":"string"},{"internalType":"string","name":"key","type":"string"}],"name":"getState","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"view","type":"function"}
]`
