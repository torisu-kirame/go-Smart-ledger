// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title LedgerAnchor — anchors Smart Ledger Merkle roots on an EVM chain (F29).
contract LedgerAnchor {
    event LedgerSealed(
        bytes32 indexed ledgerId,
        bytes32 merkleRoot,
        uint64 seqFrom,
        uint64 seqTo,
        address indexed anchoredBy
    );

    mapping(bytes32 => bytes32) public latestRoot;

    function anchor(bytes32 ledgerId, bytes32 merkleRoot, uint64 seqFrom, uint64 seqTo) external {
        latestRoot[ledgerId] = merkleRoot;
        emit LedgerSealed(ledgerId, merkleRoot, seqFrom, seqTo, msg.sender);
    }
}
