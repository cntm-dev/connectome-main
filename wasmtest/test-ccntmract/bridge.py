from cntmology.interop.Ontology.Ccntmract import Migrate

OntCversion = '2.0.0'

from cntmology.interop.Ontology.Runtime import Base58ToAddress
from cntmology.builtins import concat
from cntmology.interop.System.Action import RegisterAction
from cntmology.interop.System.App import DynamicAppCall
from cntmology.interop.System.ExecutionEngine import GetExecutingScriptHash
from cntmology.interop.System.Runtime import CheckWitness, Log
from cntmology.libcntm import bytearray_reverse, AddressFromVmCode
from cntmology.interop.System.Storage import GetCcntmext, Put, Get
from cntmology.interop.Ontology.Native import Invoke

Oep4ToErc20Event = RegisterAction("deposit", "cntm_acct", "eth_acct", "amount", "cntm_token_address", "eth_token_address")
Erc20ToOep4Event = RegisterAction("withdraw", "eth_acct", "cntm_acct", "amount", "cntm_token_address",
                                  "eth_token_address")
TestEvent = RegisterAction("testEvent", "eth_acct")

TRANSFER_ID = bytearray(b'\xa9\x05\x9c\xbb')
TRANSFER_FROM_ID = bytearray(b'\x23\xb8\x72\xdd')
BALANCEOF_ID = bytearray(b'\x70\xa0\x82\x31')

KEY_cntm_TOKEN_ARR = bytearray(b'\x01')
KEY_ETH_TOKEN_ARR = bytearray(b'\x02')

Admin = Base58ToAddress("ARGK44mXXZfU6vcdSfFKMzjaabWxyog1qb")
cntm_SYSTEM_CcntmRACT_ADDRESS = bytearray(
    b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff')
ZERO_ADDRESS = bytearray(b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00')

EVM_INVOKE_METHOD = "evmInvoke"

ctx = GetCcntmext()
ethVersion = 0


def Main(operation, args):
    if operation == 'init':
        assert (len(args) == 2)
        return init(args[0], args[1])
    if operation == "migrate":
        assert (len(args) == 6)
        return migrate(args[0], args[1], args[2], args[3], args[4], args[5])
    if operation == 'get_oep4_address':
        return get_oep4_address()
    if operation == 'get_erc20_address':
        return get_erc20_address()
    if operation == 'oep4ToErc20':
        assert (len(args) == 3)
        return oep4ToErc20(args[0], args[1], args[2])
    if operation == 'erc20ToOep4':
        assert (len(args) == 3)
        return erc20ToOep4(args[0], args[1], args[2])
    return False


def init(cntm_token_addr, eth_token_addr):
    """
    storage the cntm token address and the eth token address
    :param cntm_token_addr: the cntmology oep4 token address
    :param eth_token_addr: the ethereum erc20 token address
    :return: True means success, False or raising exception means failure.
    """
    assert (CheckWitness(Admin))
    assert (len(cntm_token_addr) == 20)
    assert (len(eth_token_addr) == 20)
    assert (cntm_token_addr != ZERO_ADDRESS)
    assert (eth_token_addr != ZERO_ADDRESS)
    Put(GetCcntmext(), KEY_cntm_TOKEN_ARR, cntm_token_addr)
    Put(GetCcntmext(), KEY_ETH_TOKEN_ARR, eth_token_addr)
    return True


def migrate(code, name, version, author, email, description):
    """
    storage the cntm token address and the eth token address
    :param cntm_token_addr: the cntmology oep4 token address
    :param eth_token_addr: the ethereum erc20 token address
    :return: True means success, False or raising exception means failure.
    """
    assert (CheckWitness(Admin))
    oep4_addr = get_oep4_address()
    erc20_address = get_erc20_address()
    success = Migrate(code, True, name, version, author, email, description)
    new_addr = AddressFromVmCode(code)
    assert (new_addr != ZERO_ADDRESS)
    new_addr = bytearray_reverse(new_addr)
    TestEvent(new_addr)
    this = GetExecutingScriptHash()
    amount = oep4BalanceOf(oep4_addr, this)
    if amount > 0:
        assert (DynamicAppCall(oep4_addr, "transfer", [this, new_addr, amount]))
    balance = erc20BalanceOf(erc20_address, this)
    if balance > 0:
        transferData = genEthTransferData(new_addr, balance)
        assert (Invoke(ethVersion, cntm_SYSTEM_CcntmRACT_ADDRESS, EVM_INVOKE_METHOD, state(erc20_address, transferData)))
    return True


def oep4ToErc20(cntm_acct, eth_acct, amount):
    """
    deposit amount of tokens from cntmology to ethereum
    :param cntm_acct: the cntmology account from which the amount of tokens will be transferred
    :param eth_acct: the ethereum account to which the amount of tokens will be transferred
    :param amount: the amount of the tokens to be deposited, >= 0
    :param token_addr: the token address
    :return: True means success, False or raising exception means failure.
    """
    assert (len(cntm_acct) == 20)
    assert (len(eth_acct) == 20)
    assert (CheckWitness(cntm_acct))
    assert (amount > 0)
    oep4_token_address = get_oep4_address()
    assert (len(oep4_token_address) == 20)
    erc20_token_address = get_erc20_address()
    assert (len(erc20_token_address) == 20)

    this = GetExecutingScriptHash()
    before = oep4BalanceOf(oep4_token_address, this)
    assert (DynamicAppCall(oep4_token_address, "transfer", [cntm_acct, this, amount]))
    after = oep4BalanceOf(oep4_token_address, this)
    assert (after >= before)
    delta = after - before
    if delta > 0:
        transferData = genEthTransferData(eth_acct, delta)
        assert (Invoke(ethVersion, cntm_SYSTEM_CcntmRACT_ADDRESS, EVM_INVOKE_METHOD, state(erc20_token_address, transferData)))
    Oep4ToErc20Event(cntm_acct, eth_acct, amount, oep4_token_address, erc20_token_address)
    return True


def oep4BalanceOf(oep4_addr, acct):
    return DynamicAppCall(oep4_addr, "balanceOf", [acct])


def erc20ToOep4(eth_acct, cntm_acct, amount):
    """
    withdraw amount of tokens from ethereum to cntmology
    :param cntm_addr: the cntmology account to which the amount of tokens will be transferred
    :param eth_acct: the ethereum account from which the amount of tokens will be transferred
    :param amount: the amount of the tokens to be withdrawed, >= 0
    :return: True means success, False or raising exception means failure.
    """
    assert (len(cntm_acct) == 20)
    assert (len(eth_acct) == 20)
    assert (CheckWitness(eth_acct))
    assert (amount > 0)
    oep4_token_address = get_oep4_address()
    assert (len(oep4_token_address) == 20)
    erc20_token_address = get_erc20_address()
    assert (len(erc20_token_address) == 20)

    this = GetExecutingScriptHash()
    before = erc20BalanceOf(erc20_token_address, this)
    tranferFromData = genEthTransferFromData(eth_acct, this, amount)
    assert (
        Invoke(ethVersion, cntm_SYSTEM_CcntmRACT_ADDRESS, EVM_INVOKE_METHOD, state(erc20_token_address, tranferFromData)))
    after = erc20BalanceOf(erc20_token_address, this)
    assert (after >= before)
    delta = after - before
    if delta > 0:
        assert (DynamicAppCall(oep4_token_address, "transfer", [this, cntm_acct, delta]))
    Erc20ToOep4Event(eth_acct, cntm_acct, amount, oep4_token_address, erc20_token_address)
    return True


def erc20BalanceOf(erc20_addr, ethAcct):
    balanceData = genEthBalanceOfData(ethAcct)
    res = Invoke(ethVersion, cntm_SYSTEM_CcntmRACT_ADDRESS, EVM_INVOKE_METHOD, state(erc20_addr, balanceData))
    return bytearray_reverse(res)


def get_oep4_address():
    return Get(GetCcntmext(), KEY_cntm_TOKEN_ARR)


def get_erc20_address():
    return Get(GetCcntmext(), KEY_ETH_TOKEN_ARR)


def genEthTransferData(to, amount):
    return concat(concat(TRANSFER_ID, formatAddr(to)), formatAmount(amount))


def genEthBalanceOfData(addr):
    return concat(BALANCEOF_ID, formatAddr(addr))


def genEthTransferFromData(from_acct, to_acct, amount):
    data = concat(TRANSFER_FROM_ID, formatAddr(from_acct))
    data = concat(data, formatAddr(to_acct))
    return concat(data, formatAmount(amount))


def formatAmount(amount):
    data = bytearray(amount)
    data = bytearray_reverse(data)
    prefix = bytearray(b'\x00')
    data_len = len(data)
    assert (data_len <= 32)
    for index in range(32 - data_len):
        data = concat(prefix, data)
    return data


def formatAddr(addr):
    assert (len(addr) == 20)
    prefix = bytearray(b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00')
    return concat(prefix, addr)