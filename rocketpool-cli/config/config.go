package config

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/config"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Configure the Rocket Pool service
func configureService(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Load global config
    globalConfig, err := rp.LoadGlobalConfig()
    if err != nil {
        return err
    }

    // Initialize user config
    userConfig := config.RocketPoolConfig{}

    // Configure chains
    if err := configureChain(&(globalConfig.Chains.Eth1), &(userConfig.Chains.Eth1), "Eth 1.0"); err != nil {
        return err
    }
    if err := configureChain(&(globalConfig.Chains.Eth2), &(userConfig.Chains.Eth2), "Eth 2.0"); err != nil {
        return err
    }

    // Save user config
    if err := rp.SaveUserConfig(userConfig); err != nil {
        return err
    }

    // Log & return
    fmt.Println("Done! Run 'rocketpool service start' to apply new configuration settings.")
    return nil

}


// Configure a chain
func configureChain(globalChain, userChain *config.Chain, chainName string) error {

    // Check client options
    if len(globalChain.Client.Options) == 0 {
        return fmt.Errorf("There are no available %s client options", chainName)
    }

    // Prompt for client
    clientOptions := make([]string, len(globalChain.Client.Options))
    for oi, option := range globalChain.Client.Options {
        clientOptions[oi] = option.Name
    }
    selected, clientName := cliutils.Select(fmt.Sprintf("Which %s client would you like to run?", chainName), clientOptions)

    // Set selected client
    globalChain.Client.Selected = globalChain.Client.Options[selected].ID
    userChain.Client.Selected = globalChain.Client.Options[selected].ID

    // Log
    fmt.Printf("%s %s client selected.\n", clientName, chainName)
    fmt.Println("")

    // Prompt for params
    params := []config.UserParam{}
    for _, param := range globalChain.GetSelectedClient().Params {

        // Get expected param format
        var expectedFormat string
        if param.Regex != "" {
            expectedFormat = param.Regex
        } else if param.Required {
            expectedFormat = "^.+$"
        } else {
            expectedFormat = "^.*$"
        }

        // Optional field text
        optionalLabel := ""
        if !param.Required {
            optionalLabel = " (leave blank for none)"
        }

        // Prompt for value
        value := cliutils.Prompt(fmt.Sprintf("Please enter the %s%s", param.Name, optionalLabel), expectedFormat, fmt.Sprintf("Invalid %s", param.Name))

        // Add param
        params = append(params, config.UserParam{
            Env: param.Env,
            Value: value,
        })

    }
    userChain.Client.Params = params

    // Return
    return nil

}

