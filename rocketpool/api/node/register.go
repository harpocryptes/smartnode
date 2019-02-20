package node

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Register the node with Rocket Pool
func registerNode(c *cli.Context) error {

    // Command setup
    if message, err := setup(c, []string{"rocketNodeAPI", "rocketNodeSettings"}, []string{}, true); message != "" {
        fmt.Println(message)
        return nil
    } else if err != nil {
        return err
    }

    // Status channels
    successChannel := make(chan bool)
    messageChannel := make(chan string)
    errorChannel := make(chan error)

    // Check if node is already registered (contract exists)
    go (func() {
        nodeContractAddress := new(common.Address)
        if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address); err != nil {
            errorChannel <- errors.New("Error checking node registration: " + err.Error())
        } else if !bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
            messageChannel <- fmt.Sprintf("Node already registered with contract at %s", nodeContractAddress.Hex())
        } else {
            successChannel <- true
        }
    })()

    // Check node registrations are enabled
    go (func() {
        registrationsAllowed := new(bool)
        if err := cm.Contracts["rocketNodeSettings"].Call(nil, registrationsAllowed, "getNewAllowed"); err != nil {
            errorChannel <- errors.New("Error checking node registrations enabled status: " + err.Error())
        } else if !*registrationsAllowed {
            messageChannel <- "Node registrations are currently disabled in Rocket Pool"
        } else {
            successChannel <- true
        }
    })()

    // Check node account ether balance
    go (func() {

        // Balance data channels
        minEtherBalanceChannel := make(chan *big.Int)
        etherBalanceChannel := make(chan *big.Int)
        balanceErrorChannel := make(chan error)

        // Get min required node account ether balance
        go (func() {
            minNodeAccountEtherBalanceWei := new(*big.Int)
            if err := cm.Contracts["rocketNodeSettings"].Call(nil, minNodeAccountEtherBalanceWei, "getEtherMin"); err != nil {
                balanceErrorChannel <- errors.New("Error retrieving minimum ether requirement: " + err.Error())
            } else {
                minEtherBalanceChannel <- *minNodeAccountEtherBalanceWei
            }
        })()

        // Get node account ether balance
        go (func() {
            if nodeAccountEtherBalanceWei, err := client.BalanceAt(context.Background(), am.GetNodeAccount().Address, nil); err != nil {
                balanceErrorChannel <- errors.New("Error retrieving node account balance: " + err.Error())
            } else {
                etherBalanceChannel <- nodeAccountEtherBalanceWei
            }
        })()

        // Receive balance data
        var minNodeAccountEtherBalanceWei *big.Int
        var nodeAccountEtherBalanceWei *big.Int
        for received := 0; received < 2; {
            select {
                case minNodeAccountEtherBalanceWei = <-minEtherBalanceChannel:
                    received++
                case nodeAccountEtherBalanceWei = <-etherBalanceChannel:
                    received++
                case err := <-balanceErrorChannel:
                    errorChannel <- err
                    return
            }
        }

        // Check node account ether balance
        if nodeAccountEtherBalanceWei.Cmp(minNodeAccountEtherBalanceWei) < 0 {
            messageChannel <- fmt.Sprintf("Node account requires a minimum balance of %.2f ETH to register", eth.WeiToEth(minNodeAccountEtherBalanceWei))
        } else {
            successChannel <- true
        }

    })()

    // Receive status
    for received := 0; received < 3; {
        select {
            case <-successChannel:
                received++
            case msg := <-messageChannel:
                fmt.Println(msg)
                return nil
            case err := <-errorChannel:
                return err
        }
    }

    // Prompt user for timezone
    timezone := promptTimezone()

    // Register node
    if txor, err := am.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        if _, err := cm.Contracts["rocketNodeAPI"].Transact(txor, "add", timezone); err != nil {
            return errors.New("Error registering node: " + err.Error())
        }
    }

    // Get node contract address
    nodeContractAddress := new(common.Address)
    if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address); err != nil {
        return errors.New("Error retrieving node contract address: " + err.Error())
    }

    // Log & return
    fmt.Println("Node registered successfully with contract at", nodeContractAddress.Hex())
    return nil

}
