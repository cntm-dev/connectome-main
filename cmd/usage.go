/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */
package cmd

import "fmt"

func showAssetHelp() {
	var assetHelp = `
   Name:
      cntmology asset                       asset operation

   Usage:
      cntmology asset [command options] [args]

   Description:
      With this command, you can ccntmrol assert through transaction.

   Command:
      transfer
         --caddr     value                 smart ccntmract address
         --from      value                 wallet address base58, which will transfer from
         --to        value                 wallet address base58, which will transfer to
         --value     value                 how much asset will be transfered
         --password  value                 use password who transfer from

      status
         --hash     value                  transfer transaction hash
`
	fmt.Println(assetHelp)
}

func showQueryAssetTransferHelp() {
	var queryAssetTransferHelp = `
   Name:
      cntmology asset query              asset transfer resule query

   Usage:
      cntmology asset query [command options] [args]

   Description:
      With this command, you can query transfer assert status.

   Command:
      --hash     value                    transfer transaction hash
`
	fmt.Println(queryAssetTransferHelp)
}

func showAssetTransferHelp() {
	var assetTransferHelp = `
   Name:
      cntmology asset transfer              asset transfer

   Usage:
      cntmology asset transfer [command options] [args]

   Description:
      With this command, you can transfer assert through transaction.

   Command:
      --caddr     value                    smart ccntmract address
      --from      value                    wallet address base58, which will transfer from
      --to        value                    wallet address base58, which will transfer to
      --value     value                    how much asset will be transfered
      --password  value                    use password who transfer from
`
	fmt.Println(assetTransferHelp)
}

func showCcntmractHelp() {
	var ccntmractUsingHelp = `
   Name:
      cntmology ccntmract      deploy or invoke a smart ccntmract by this command
   Usage:
      cntmology ccntmract [command options] [args]

   Description:
      With this command, you can invoke a smart ccntmract

   Command:
     invoke
       --caddr      value               smart ccntmract address that will be invoke
       --params     value               params will be  
			
     deploy
       --type       value               ccntmract type ,value: 1 (NEOVM) | 2 (WASM)
       --store      value               does this ccntmract will be stored, value: true or false
       --code       value               directory of smart ccntmract that will be deployed
       --cname      value               ccntmract name that will be deployed
       --cversion   value               ccntmract version which will be deployed
       --author     value               owner of deployed smart ccntmract
       --email      value               owner email who deploy the smart ccntmract
       --desc       value               ccntmract description when deploy one
`
	fmt.Println(ccntmractUsingHelp)
}

func showDeployHelp() {
	var deployHelp = `
   Name:
      cntmology ccntmract deploy        deploy a smart ccntmract by this command
   Usage:
      cntmology ccntmract deploy [command options] [args]

   Description:
      With this command, you can deploy a smart ccntmract

   Command:
      --type       value              ccntmract type ,value: 1 (NEOVM) | 2 (WASM)
      --store      value              does this ccntmract will be stored, value: true or false
      --code       value              directory of smart ccntmract that will be deployed
      --cname      value              ccntmract name that will be deployed
      --cversion   value              ccntmract version which will be deployed
      --author     value              owner of deployed smart ccntmract
      --email      value              owner email who deploy the smart ccntmract
      --desc       value              ccntmract description when deploy one
`
	fmt.Println(deployHelp)
}
func showInvokeHelp() {
	var invokeHelp = `
   Name:
      cntmology ccntmract invoke          invoke a smart ccntmract by this command
   Usage:
      cntmology ccntmract invoke [command options] [args]

   Description:
      With this command, you can invoke a smart ccntmract

   Command:
      --caddr      value                smart ccntmract address that will be invoke
      --params     value                params will be
`
	fmt.Println(invokeHelp)
}

func showInfoHelp() {
	var infoHelp = `
   Name:
      cntmology info                    Show blockchain information

   Usage:
      cntmology info [command options] [args]

   Description:
      With cntmology info, you can look up blocks, transactions, etc.

   Command:
      version

      block
         --hash value                  block hash value
         --height value                block height value

      tx
         --hash value                  transaction hash value

`
	fmt.Println(infoHelp)
}

func showVersionInfoHelp() {
	var versionInfoHelp = `
   Name:
      cntmology info version            Show cntmology node version

   Usage:
      cntmology info version

   Description:
      With this command, you can look up the cntmology node version.

`
	fmt.Println(versionInfoHelp)
}

func showBlockInfoHelp() {
	var blockInfoHelp = `
   Name:
      cntmology info block             Show blockchain information

   Usage:
      cntmology info block [command options] [args]

   Description:
      With this command, you can look up block information.

   Options:
      --hash value                    block hash value
      --height value                  block height value
`
	fmt.Println(blockInfoHelp)
}

func showTxInfoHelp() {
	var txInfoHelp = `
   Name:
      cntmology info tx               Show transaction information

   Usage:
      cntmology info tx [command options] [args]

   Description:
      With this command, you can look up transaction information.

   Options:
      --hash value                   transaction hash value

`
	fmt.Println(txInfoHelp)
}

func showSettingHelp() {
	var settingHelp = `
   Name:
      cntmology set                       Show blockchain information

   Usage:
      cntmology set [command options] [args]

   Description:
      With cntmology set, you can configure the node.

   Command:
      --debuglevel value                 debug level(0~6) will be set
      --consensus value                  [ on / off ]
`
	fmt.Println(settingHelp)
}

func showWalletHelp() {
	var walletHelp = `
   Name:
      cntmology wallet                  User wallet operation

   Usage:
      cntmology wallet [command options] [args]

   Description:
      With cntmology wallet, you could ccntmrol your account.

   Command:
      create
      --name value                     wallet name
      show
      --name value                     wallet name (default: wallet.dat)
      balance
      --name value                     wallet name (default: wallet.dat)
`
	fmt.Println(walletHelp)
}
