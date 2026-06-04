// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title Smart Ledger registry on FISCO BCOS (EVM compatible)
/// @notice v0.14 — skeleton for ledger meta, events, members, pending approvals, accounting JSON blobs.
/// @dev Deploy on FISCO BCOS group; ledger-api calls via go-sdk (see docs/fisco-bcos-migration.md).
contract LedgerRegistry {
    struct LedgerMeta {
        string name;
        string bookkeepingMode; // simple | professional
        string ledgerType; // private | multi
        uint64 latestSeq;
        bytes32 latestRoot;
        string anchorStatus;
    }

    struct EventRecord {
        uint64 seq;
        string eventType;
        bytes32 hash;
        string signerId;
        bytes payload;
    }

    mapping(string => LedgerMeta) public ledgers;
    mapping(string => mapping(uint64 => EventRecord)) public events;
    mapping(string => mapping(string => bool)) public members;
    mapping(string => mapping(string => bytes)) public stateBlobs; // accounting / collaboration JSON by key

    event LedgerCreated(string indexed ledgerId, string creatorId);
    event EventAppended(string indexed ledgerId, uint64 seq, string eventType);
    event MemberUpdated(string indexed ledgerId, string memberId, bool active);

    function createLedger(
        string calldata ledgerId,
        string calldata creatorId,
        string calldata name,
        string calldata bookkeepingMode,
        string calldata ledgerType
    ) external {
        require(bytes(ledgers[ledgerId].name).length == 0, "exists");
        ledgers[ledgerId] = LedgerMeta({
            name: name,
            bookkeepingMode: bookkeepingMode,
            ledgerType: ledgerType,
            latestSeq: 0,
            latestRoot: bytes32(0),
            anchorStatus: "pending"
        });
        members[ledgerId][creatorId] = true;
        emit LedgerCreated(ledgerId, creatorId);
    }

    function appendEvent(
        string calldata ledgerId,
        string calldata eventType,
        bytes32 hash,
        string calldata signerId,
        bytes calldata payload
    ) external returns (uint64 seq) {
        require(bytes(ledgers[ledgerId].name).length > 0, "no ledger");
        require(members[ledgerId][signerId], "not member");
        seq = ledgers[ledgerId].latestSeq + 1;
        events[ledgerId][seq] = EventRecord({
            seq: seq,
            eventType: eventType,
            hash: hash,
            signerId: signerId,
            payload: payload
        });
        ledgers[ledgerId].latestSeq = seq;
        emit EventAppended(ledgerId, seq, eventType);
    }

    function setMember(string calldata ledgerId, string calldata memberId, bool active) external {
        members[ledgerId][memberId] = active;
        emit MemberUpdated(ledgerId, memberId, active);
    }

    function putState(string calldata ledgerId, string calldata key, bytes calldata value) external {
        require(members[ledgerId][msg.sender] || bytes(key).length > 0, "unauthorized");
        stateBlobs[ledgerId][key] = value;
    }

    function getState(string calldata ledgerId, string calldata key) external view returns (bytes memory) {
        return stateBlobs[ledgerId][key];
    }
}
