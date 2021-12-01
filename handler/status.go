package handler

import (
	"bytes"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/genesis"
	"github.com/hacash/core/interfaces"
)

const (
	HandShakeStatusDataSize = 32 + 1 + 1 + 2 + 2 + 8 + 32
)

type HandShakeStatusData struct {

	// network
	GenesisBlockHash fields.Hash

	// version
	BlockVersion    fields.VarUint1 // uint8
	TransactionType fields.VarUint1 // uint8
	ActionKind      fields.VarUint2 // uint16
	RepairVersion   fields.VarUint2 // uint16

	// status
	LastestBlockHeight fields.VarUint8 // uint64
	LastesBlockHash    fields.Hash
}

func (this *HandShakeStatusData) Size() uint32 {
	return HandShakeStatusDataSize
}

func (this *HandShakeStatusData) Serialize() ([]byte, error) {
	var buffer = new(bytes.Buffer)
	b1, _ := this.GenesisBlockHash.Serialize()
	b2, _ := this.BlockVersion.Serialize()
	b3, _ := this.TransactionType.Serialize()
	b4, _ := this.ActionKind.Serialize()
	b5, _ := this.RepairVersion.Serialize()
	b6, _ := this.LastestBlockHeight.Serialize()
	b7, _ := this.LastesBlockHash.Serialize()
	buffer.Write(b1)
	buffer.Write(b2)
	buffer.Write(b3)
	buffer.Write(b4)
	buffer.Write(b5)
	buffer.Write(b6)
	buffer.Write(b7)
	return buffer.Bytes(), nil
}

func (this *HandShakeStatusData) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = this.GenesisBlockHash.Parse(buf, seek)
	seek, _ = this.BlockVersion.Parse(buf, seek)
	seek, _ = this.TransactionType.Parse(buf, seek)
	seek, _ = this.ActionKind.Parse(buf, seek)
	seek, _ = this.RepairVersion.Parse(buf, seek)
	seek, _ = this.LastestBlockHeight.Parse(buf, seek)
	seek, _ = this.LastesBlockHash.Parse(buf, seek)
	return seek, nil
}

////////////////////////////////////////////////////////

func SendStatusToPeer(blockchain interfaces.BlockChain, peer interfaces.P2PMsgPeer) {

	lastblock, _, err := blockchain.GetChainEngineKernel().LatestBlock()
	if err != nil {
		panic(err)
	}

	statusData := HandShakeStatusData{
		genesis.GetGenesisBlock().Hash(),
		blocks.BlockVersion,
		blocks.TransactionType,
		blocks.ActionKind,
		blocks.RepairVersion,
		fields.VarUint8(lastblock.GetHeight()),
		lastblock.Hash(),
	}

	msgdata, _ := statusData.Serialize()
	// send
	peer.SendDataMsg(MsgTypeStatus, msgdata)
}

func GetStatus(blockchain interfaces.BlockChain, peer interfaces.P2PMsgPeer, msgbody []byte) {

	if len(msgbody) != HandShakeStatusDataSize {
		peer.Disconnect()
		return
	}
	var otherStatusObj HandShakeStatusData
	otherStatusObj.Parse(msgbody, 0)
	// check hand shake
	if bytes.Compare(genesis.GetGenesisBlock().Hash(), otherStatusObj.GenesisBlockHash) != 0 {
		fmt.Println(fmt.Errorf("Disconnect peer " + peer.Describe() + ", Genesis block hash is difference."))
		peer.Disconnect()
		return
	}
	// check version
	var isUpdate = false
	if blocks.BlockVersion < otherStatusObj.BlockVersion || blocks.TransactionType < otherStatusObj.TransactionType || blocks.ActionKind < otherStatusObj.ActionKind {
		isUpdate = true // update
	} else if blocks.BlockVersion == otherStatusObj.BlockVersion && blocks.TransactionType == otherStatusObj.TransactionType && blocks.ActionKind == otherStatusObj.ActionKind && blocks.RepairVersion < otherStatusObj.RepairVersion {
		isUpdate = true // update
	}
	if isUpdate {
		printWarning(
			"[Warning] You must update the Hacash node software form https://hacash.org\n" +
				"          If an error is reported after upgrading the new version, please try to delete all block data and resynchronize\n" +
				"【警告】 你的节点软件版本低于全网正在使用的版本，升级 Hacash 的节点软件，请访问 https://hacash.org\n" +
				"        如果升级新版本之后运行报错，请尝试删除所有区块数据后重新同步")
		peer.Disconnect()
		return
	}
	// my state
	lastblock, _, err := blockchain.GetChainEngineKernel().LatestBlock()
	if err != nil {
		panic(err)
	}
	mylastblockheight := lastblock.GetHeight()
	// fork or sync new block
	if mylastblockheight == 0 && otherStatusObj.LastestBlockHeight > 0 {
		// first sync block data
		msgParseSendRequestBlocks(peer, 1)
	} else if uint64(otherStatusObj.LastestBlockHeight) > mylastblockheight {
		// check hash fork
		msgParseSendRequestBlockHashList(peer, 4, mylastblockheight)
	}

}
