// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract SecurityDeposit {
    // Mapping of user's addresses to their deposited amount
    mapping(address => uint256) public deposits;
    // Mapping to keep track of registered sequencers
    mapping(address => bool) public isSequencer;

    // Admin address
    address public admin;
    // Constant inclusion fee
    uint256 public INCLUSION_FEE;

    // Event declarations
    event Deposit(address indexed user, uint256 amount);
    event Withdrawal(address indexed user, uint256 amount);
    event SequencerRegistered(address sequencer);
    event FeesCollected(address indexed sequencer, address indexed user, uint256 fee);
    event FeesRefunded(address indexed sequencer, address indexed user, uint256 fee);
    event InclusionFeeUpdated(uint256 newFee);

    // Constructor to set the initial inclusion fee and admin
    constructor(uint256 initialFee) {
        INCLUSION_FEE = initialFee;
        admin = msg.sender;
    }

    // Modifier to restrict access to the admin
    modifier onlyAdmin() {
        require(msg.sender == admin, "Only admin can perform this action");
        _;
    }

    // Modifier to restrict access to registered sequencers
    modifier onlySequencer() {
        require(isSequencer[msg.sender], "Only a registered sequencer can perform this action");
        _;
    }

    // Function to update the inclusion fee
    function updateInclusionFee(uint256 newFee) public onlyAdmin {
        INCLUSION_FEE = newFee;
        emit InclusionFeeUpdated(newFee);
    }

    // Function to deposit ETH into the contract
    function deposit() public payable {
        require(msg.value > 0, "Deposit amount must be greater than 0");
        deposits[msg.sender] += msg.value;
        emit Deposit(msg.sender, msg.value);
    }

    // Function to withdraw ETH from the contract
    function withdraw(uint256 amount) public {
        require(amount <= deposits[msg.sender], "Insufficient balance");
        deposits[msg.sender] -= amount;
        payable(msg.sender).transfer(amount);
        emit Withdrawal(msg.sender, amount);
    }

    // Function to register as a sequencer
    function registerSequencer() public {
        require(!isSequencer[msg.sender], "Already a sequencer");
        isSequencer[msg.sender] = true;
        emit SequencerRegistered(msg.sender);
    }

    // Function for sequencers to collect fees
    function collectFees(address[] calldata users) public onlySequencer {
        for (uint i = 0; i < users.length; i++) {
            require(deposits[users[i]] >= INCLUSION_FEE, "Insufficient balance for fee collection");
            deposits[users[i]] -= INCLUSION_FEE;
            emit FeesCollected(msg.sender, users[i], INCLUSION_FEE);
        }
    }

    // Function for sequencers to refund fees
    function refundFees(address[] calldata users) public onlySequencer {
        for (uint i = 0; i < users.length; i++) {
            // Increment the user's deposit by the inclusion fee
            deposits[users[i]] += INCLUSION_FEE;
            emit FeesRefunded(msg.sender, users[i], INCLUSION_FEE);
        }
    }
}
