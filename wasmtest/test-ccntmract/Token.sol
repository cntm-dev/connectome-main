// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.1;

import "@openzeppelin/ccntmracts/token/ERC20/ERC20.sol";
import "@openzeppelin/ccntmracts/access/Ownable.sol";

//SPDX-License-Identifier: <SPDX-License>
ccntmract WingToken is ERC20, Ownable {

    constructor()ERC20("Wing Token", "WING"){
        uint totalSupply = 500000 * (10 ** decimals());
        _mint(owner(), totalSupply);
    }

    function decimals() public pure override returns (uint8){
        return 9;
    }

    function mint(address to, uint amount) public onlyOwner() {
        _mint(to, amount);
    }

    function burn(uint amount) public {
        _burn(_msgSender(), amount);
    }
}