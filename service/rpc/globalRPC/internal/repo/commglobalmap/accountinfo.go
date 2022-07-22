package commglobalmap

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/errcode"
)

func (m *model) GetLatestAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	f := func() (interface{}, error) {
		accountInfo, err := m.GetLatestAccountInfo(ctx, accountIndex)
		if err != nil {
			return nil, err
		}
		account, err := commonAsset.FromFormatAccountInfo(accountInfo)
		if err != nil {
			return nil, errcode.ErrFromFormatAccountInfo.RefineError(fmt.Sprint(accountInfo))
		}
		return account, nil
	}
	accountInfo := &account.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), accountInfo, 10, f)
	if err != nil {
		return nil, err
	}
	account, _ := value.(*account.Account)
	res, err := commonAsset.ToFormatAccountInfo(account)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *model) SetLatestAccountInfoInToCache(ctx context.Context, accountIndex int64) error {
	accountInfo, err := m.GetLatestAccountInfo(ctx, accountIndex)
	if err != nil {
		return err
	}
	account, err := commonAsset.FromFormatAccountInfo(accountInfo)
	if err != nil {
		return err
	}
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), account, 10); err != nil {
		return err
	}
	return nil
}

func (m *model) DeleteLatestAccountInfoInCache(ctx context.Context, accountIndex int64) error {
	return m.cache.Delete(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex))
}

func (m *model) GetLatestAccountInfo(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprint("GetAccountByAccountIndex:", err.Error()))
	}
	accountInfo, err := commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		return nil, errcode.ErrToFormatAccountInfo.RefineError(err.Error())
	}
	mempoolTxs, err := m.mempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil && err != mempool.ErrNotFound {
		return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprint("GetPendingMempoolTxsByAccountIndex:", err.Error()))
	}
	for _, mempoolTx := range mempoolTxs {
		if mempoolTx.Nonce != commonConstant.NilNonce {
			accountInfo.Nonce = mempoolTx.Nonce
		}
		for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
			if mempoolTxDetail.AccountIndex != accountIndex {
				continue
			}
			switch mempoolTxDetail.AssetType {
			case commonAsset.GeneralAssetType:
				if accountInfo.AssetInfo[mempoolTxDetail.AssetId] == nil {
					accountInfo.AssetInfo[mempoolTxDetail.AssetId] = &commonAsset.AccountAsset{
						AssetId:                  mempoolTxDetail.AssetId,
						Balance:                  util.ZeroBigInt,
						LpAmount:                 util.ZeroBigInt,
						OfferCanceledOrFinalized: util.ZeroBigInt,
					}
				}
				nBalance, err := commonAsset.ComputeNewBalance(commonAsset.GeneralAssetType,
					accountInfo.AssetInfo[mempoolTxDetail.AssetId].String(), mempoolTxDetail.BalanceDelta)
				if err != nil {
					return nil, errcode.ErrComputeNewBalance.RefineError(err.Error())
				}
				accountInfo.AssetInfo[mempoolTxDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
				if err != nil {
					return nil, errcode.ErrParseAccountAsset.RefineError(err.Error())
				}
			case commonAsset.CollectionNonceAssetType:
				accountInfo.CollectionNonce, err = strconv.ParseInt(mempoolTxDetail.BalanceDelta, 10, 64)
				if err != nil {
					return nil, errcode.ErrParseInt.RefineError(mempoolTxDetail.BalanceDelta)
				}
			case commonAsset.LiquidityAssetType:
			case commonAsset.NftAssetType:
			default:
				return nil, errcode.ErrInvalidAssetType
			}
		}
	}
	accountInfo.Nonce = accountInfo.Nonce + 1
	accountInfo.CollectionNonce = accountInfo.CollectionNonce + 1
	return accountInfo, nil
}

func (m *model) GetBasicAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	f := func() (interface{}, error) {
		accountInfo, err := m.GetBasicAccountInfo(ctx, accountIndex)
		if err != nil {
			return nil, err
		}
		account, err := commonAsset.FromFormatAccountInfo(accountInfo)
		if err != nil {
			return nil, errcode.ErrFromFormatAccountInfo.RefineError(fmt.Sprint(accountInfo))
		}
		return account, nil
	}
	accountInfo := &account.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyBasicAccountByAccountIndex(accountIndex), accountInfo, 10, f)
	if err != nil {
		return nil, err
	}
	account, _ := value.(*account.Account)
	res, err := commonAsset.ToFormatAccountInfo(account)
	if err != nil {
		return nil, errcode.ErrToFormatAccountInfo.RefineError(fmt.Sprint(accountInfo))
	}
	return res, nil
}

func (m *model) GetBasicAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprint("GetAccountByAccountIndex:", err.Error()))
	}
	accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		return nil, errcode.ErrToFormatAccountInfo.RefineError(fmt.Sprint(accountInfo))
	}
	return accountInfo, nil
}