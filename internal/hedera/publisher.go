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
		_, err = hedera.NewTopicMessageSubmitTransaction().
			SetMessage(b).
			SetTopicID(topicId).
			Execute(p.hederaClient)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *HederaPublisher) Close() error {
	return p.hederaClient.Close()
}
