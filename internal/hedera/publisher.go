package hedera

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/contracts"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/message"
	logInterface "github.com/project-alvarium/provider-logging/pkg/interfaces"
	"github.com/project-alvarium/provider-logging/pkg/logging"
)

type HederaPublisher struct {
	cfg          config.HederaConfig
	logger       logInterface.Logger
	hederaClient *hedera.Client
}

func NewHederaPublisher(cfg config.HederaConfig, logger logInterface.Logger) (interfaces.StreamProvider, error) {
	var client *hedera.Client
	switch netType := cfg.NetType; netType {
	case contracts.Mainnet:
		client = hedera.ClientForMainnet()
	case contracts.Testnet:
		client = hedera.ClientForTestnet()
	case contracts.Previewnet:
		client = hedera.ClientForPreviewnet()
	default:
		return nil, errors.New("nettype not valid")
	}

	accountId, err := hedera.AccountIDFromString(cfg.AccountId)
	if err != nil {
		return nil, err
	}

	privateKey, err := hedera.PrivateKeyFromString(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	client.SetOperator(accountId, privateKey)

	p := HederaPublisher{
		cfg:          cfg,
		logger:       logger,
		hederaClient: client,
	}
	return &p, nil
}

// hedera client implicitly connects to the hedera net.
// no need for manual initiation.
func (p *HederaPublisher) Connect() error {
	return nil
}

func (p *HederaPublisher) Publish(msg message.PublishWrapper) error {
	b, _ := json.Marshal(msg)

	// publish to all topic IDs
	for _, topic := range p.cfg.Topics {
		p.logger.Write(logging.DebugLevel, fmt.Sprintf("attempting publish, topic %s %s", topic, string(b)))
		topicId, err := hedera.TopicIDFromString(topic)
		if err != nil {
			return err
		}

		// submit message to consensus service
		transaction, err := hedera.NewTopicMessageSubmitTransaction().
			SetMessage(b).
			SetTopicID(topicId).
			Execute(p.hederaClient)
		if err != nil {
			return err
		}

		contractId, err := hedera.ContractIDFromString(p.cfg.ContractId)
		if err != nil {
			return err
		}
		publishParams := hedera.NewContractFunctionParameters().
			AddString(transaction.TransactionID.String())

		// call the `publish` smart contract function which will record the
		// previous transaction's ID in the ledger to be used to calculate
		// the fees in bulk.
		//
		// A transaction is used here instead of a query because queries
		// usually return a `BUSY` error which is due to high traffic when
		// using the testnet or previewnet
		tr, err := hedera.NewContractExecuteTransaction().
			SetContractID(contractId).
			SetGas(1000000).
			SetFunction("publish", publishParams).
			Execute(p.hederaClient)

		// At the moment, the is no cleaning up if the annotation is published to
		// the consensus service but the smart contract call fails
		if err != nil {
			return err
		}

		// Confirmation that the transaction was successful
		receipt, err := tr.GetReceipt(p.hederaClient)
		if err != nil {
			return err
		}

		if receipt.Status != hedera.StatusSuccess {
			errMsg := fmt.Sprintf(
				"Smart contract 'publish' method invocation failed with status: %s",
				receipt.Status,
			)
			p.logger.Error(errMsg)
			return errors.New(errMsg)
		}

	}

	return nil
}

func (p *HederaPublisher) Close() error {
	return p.hederaClient.Close()
}
