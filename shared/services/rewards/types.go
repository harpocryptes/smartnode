package rewards

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/wealdtech/go-merkletree"
)

// Interface for version-agnostic minipool performance
type IMinipoolPerformanceFile interface {
	// Serialize a minipool performance file into bytes
	Serialize() ([]byte, error)

	// Serialize a minipool performance file into bytes designed for human readability
	SerializeHuman() ([]byte, error)
}

// Interface for version-agnostic rewards files
type IRewardsFile interface {
	// Serialize a rewards file into bytes
	Serialize() ([]byte, error)

	// Deserialize a rewards file from bytes
	Deserialize([]byte) error

	// Get the rewards file's header
	GetHeader() *RewardsFileHeader

	// Get info about a node's rewards
	GetNodeRewardsInfo(address common.Address) (INodeRewardsInfo, bool)

	// Gets the minipool performance file corresponding to this rewards file
	GetMinipoolPerformanceFile() IMinipoolPerformanceFile

	// Sets the CID of the minipool performance file corresponding to this rewards file
	SetMinipoolPerformanceFileCID(cid string)
}

// Rewards per network
type NetworkRewardsInfo struct {
	CollateralRpl    *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *QuotedBigInt `json:"smoothingPoolEth"`
}

// Total cumulative rewards for an interval
type TotalRewards struct {
	ProtocolDaoRpl               *QuotedBigInt `json:"protocolDaoRpl"`
	TotalCollateralRpl           *QuotedBigInt `json:"totalCollateralRpl"`
	TotalOracleDaoRpl            *QuotedBigInt `json:"totalOracleDaoRpl"`
	TotalSmoothingPoolEth        *QuotedBigInt `json:"totalSmoothingPoolEth"`
	PoolStakerSmoothingPoolEth   *QuotedBigInt `json:"poolStakerSmoothingPoolEth"`
	NodeOperatorSmoothingPoolEth *QuotedBigInt `json:"nodeOperatorSmoothingPoolEth"`
}

// Interface for version-agnostic node operator rewards
type INodeRewardsInfo interface {
	GetRewardNetwork() uint64
	GetCollateralRpl() *QuotedBigInt
	GetOracleDaoRpl() *QuotedBigInt
	GetSmoothingPoolEth() *QuotedBigInt
	GetMerkleProof() ([]common.Hash, error)
}

// Basic node operator rewards
type NodeRewardsInfo struct {
	RewardNetwork    uint64        `json:"rewardNetwork"`
	CollateralRpl    *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *QuotedBigInt `json:"smoothingPoolEth"`
	MerkleData       []byte        `json:"-"`
	MerkleProof      []string      `json:"merkleProof"`
}

func (i *NodeRewardsInfo) GetRewardNetwork() uint64 {
	return i.RewardNetwork
}
func (i *NodeRewardsInfo) GetCollateralRpl() *QuotedBigInt {
	return i.CollateralRpl
}
func (i *NodeRewardsInfo) GetOracleDaoRpl() *QuotedBigInt {
	return i.OracleDaoRpl
}
func (i *NodeRewardsInfo) GetSmoothingPoolEth() *QuotedBigInt {
	return i.SmoothingPoolEth
}
func (n *NodeRewardsInfo) GetMerkleProof() ([]common.Hash, error) {
	proof := []common.Hash{}
	for _, proofLevel := range n.MerkleProof {
		proof = append(proof, common.HexToHash(proofLevel))
	}
	return proof, nil
}

// General version-agnostic information about a rewards file
type RewardsFileHeader struct {
	// Serialized fields
	RewardsFileVersion         uint64                         `json:"rewardsFileVersion"`
	RulesetVersion             uint64                         `json:"rulesetVersion,omitempty"`
	Index                      uint64                         `json:"index"`
	Network                    string                         `json:"network"`
	StartTime                  time.Time                      `json:"startTime,omitempty"`
	EndTime                    time.Time                      `json:"endTime"`
	ConsensusStartBlock        uint64                         `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock          uint64                         `json:"consensusEndBlock"`
	ExecutionStartBlock        uint64                         `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock          uint64                         `json:"executionEndBlock"`
	IntervalsPassed            uint64                         `json:"intervalsPassed"`
	MerkleRoot                 string                         `json:"merkleRoot,omitempty"`
	MinipoolPerformanceFileCID string                         `json:"minipoolPerformanceFileCid,omitempty"`
	TotalRewards               *TotalRewards                  `json:"totalRewards"`
	NetworkRewards             map[uint64]*NetworkRewardsInfo `json:"networkRewards"`

	// Non-serialized fields
	MerkleTree          *merkletree.MerkleTree    `json:"-"`
	InvalidNetworkNodes map[common.Address]uint64 `json:"-"`
}

// Information about an interval
type IntervalInfo struct {
	Index                  uint64        `json:"index"`
	TreeFilePath           string        `json:"treeFilePath"`
	TreeFileExists         bool          `json:"treeFileExists"`
	MerkleRootValid        bool          `json:"merkleRootValid"`
	CID                    string        `json:"cid"`
	StartTime              time.Time     `json:"startTime"`
	EndTime                time.Time     `json:"endTime"`
	NodeExists             bool          `json:"nodeExists"`
	CollateralRplAmount    *QuotedBigInt `json:"collateralRplAmount"`
	ODaoRplAmount          *QuotedBigInt `json:"oDaoRplAmount"`
	SmoothingPoolEthAmount *QuotedBigInt `json:"smoothingPoolEthAmount"`
	MerkleProof            []common.Hash `json:"merkleProof"`
}

type MinipoolInfo struct {
	Address                 common.Address
	ValidatorPubkey         types.ValidatorPubkey
	ValidatorIndex          string
	NodeAddress             common.Address
	NodeIndex               uint64
	Fee                     *big.Int
	MissedAttestations      uint64
	GoodAttestations        uint64
	MinipoolShare           *big.Int
	MissingAttestationSlots map[uint64]bool
	WasActive               bool
	StartSlot               uint64
	EndSlot                 uint64
	AttestationScore        *big.Int
	CompletedAttestations   map[uint64]bool
	AttestationCount        int
}

type IntervalDutiesInfo struct {
	Index uint64
	Slots map[uint64]*SlotInfo
}

type SlotInfo struct {
	Index      uint64
	Committees map[uint64]*CommitteeInfo
}

type CommitteeInfo struct {
	Index     uint64
	Positions map[int]*MinipoolInfo
}

// Details about a node for the Smoothing Pool
type NodeSmoothingDetails struct {
	Address          common.Address
	IsEligible       bool
	IsOptedIn        bool
	StatusChangeTime time.Time
	Minipools        []*MinipoolInfo
	EligibleSeconds  *big.Int
	StartSlot        uint64
	EndSlot          uint64
	SmoothingPoolEth *big.Int
	RewardsNetwork   uint64

	// v2 Fields
	OptInTime  time.Time
	OptOutTime time.Time
}

type QuotedBigInt struct {
	big.Int
}

func NewQuotedBigInt(x int64) *QuotedBigInt {
	q := QuotedBigInt{}
	native := big.NewInt(x)
	q.Int = *native
	return &q
}

func (b *QuotedBigInt) MarshalJSON() ([]byte, error) {
	return []byte("\"" + b.String() + "\""), nil
}

func (b *QuotedBigInt) UnmarshalJSON(p []byte) error {
	strippedString := strings.Trim(string(p), "\"")
	nativeInt, success := big.NewInt(0).SetString(strippedString, 0)
	if !success {
		return fmt.Errorf("%s is not a valid big integer", strippedString)
	}

	b.Int = *nativeInt
	return nil
}
