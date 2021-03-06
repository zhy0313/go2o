/**
 * Copyright 2015 @ z3q.net.
 * name : shop_manager.go
 * author : jarryliu
 * date : 2016-05-28 12:13
 * description :
 * history :
 */
package shop

import (
	"errors"
	"go2o/core/domain/interface/enum"
	"go2o/core/domain/interface/merchant"
	"go2o/core/domain/interface/merchant/shop"
	"go2o/core/domain/interface/valueobject"
	"strings"
	"time"
)

var _ shop.IShopManager = new(shopManagerImpl)

type shopManagerImpl struct {
	_merchant merchant.IMerchant
	_rep      shop.IShopRep
	valueRep  valueobject.IValueRep
}

func NewShopManagerImpl(m merchant.IMerchant, rep shop.IShopRep,
	valueRep valueobject.IValueRep) shop.IShopManager {
	return &shopManagerImpl{
		_merchant: m,
		_rep:      rep,
		valueRep:  valueRep,
	}
}

// 新建商店
func (s *shopManagerImpl) CreateShop(v *shop.Shop) shop.IShop {
	v.CreateTime = time.Now().Unix()
	v.MerchantId = s._merchant.GetAggregateRootId()
	return newShop(s, v, s._rep, s.valueRep)
}

// 获取所有商店
func (s *shopManagerImpl) GetShops() []shop.IShop {
	shopList := s._rep.GetShopsOfMerchant(s._merchant.GetAggregateRootId())
	shops := make([]shop.IShop, len(shopList))
	for i, v := range shopList {
		v2 := v
		shops[i] = s.CreateShop(&v2)
	}
	return shops
}

// 根据名称获取商店
func (s *shopManagerImpl) GetShopByName(name string) shop.IShop {
	name = strings.TrimSpace(name)
	for _, v := range s.GetShops() {
		if strings.TrimSpace(v.GetValue().Name) == name {
			return v
		}
	}
	return nil
}

// 获取营业中的商店
func (s *shopManagerImpl) GetBusinessInShops() []shop.IShop {
	list := make([]shop.IShop, 0)
	for _, v := range s.GetShops() {
		if v.GetValue().State == enum.ShopBusinessIn {
			list = append(list, v)
		}
	}
	return list
}

// 获取商店
func (s *shopManagerImpl) GetShop(shopId int) shop.IShop {
	shops := s.GetShops()
	for _, v := range shops {
		time.Sleep(time.Microsecond * 5)
		if v.GetValue().Id == shopId {
			return v
		}
	}
	return nil
}

// 获取商铺
func (s *shopManagerImpl) GetOnlineShop() shop.IShop {
	for _, v := range s.GetShops() {
		if v.Type() == shop.TypeOnlineShop {
			return v
		}
	}
	return nil
}

// 删除门店
func (s *shopManagerImpl) DeleteShop(shopId int) error {
	//todo : 检测订单数量
	mchId := s._merchant.GetAggregateRootId()
	sp := s.GetShop(shopId)
	if sp != nil {
		switch sp.Type() {
		case shop.TypeOfflineShop:
			return s.deleteOfflineShop(mchId, sp)
		case shop.TypeOnlineShop:
			return s.deleteOnlineShop(mchId, sp)
		}
	}
	return nil
}

func (s *shopManagerImpl) deleteOnlineShop(mchId int, sp shop.IShop) error {
	return errors.New("暂不支持删除线上商店")
	shopId := sp.GetDomainId()
	err := s._rep.DeleteOnlineShop(mchId, shopId)
	return err
}

func (s *shopManagerImpl) deleteOfflineShop(mchId int, sp shop.IShop) error {
	shopId := sp.GetDomainId()
	err := s._rep.DeleteOfflineShop(mchId, shopId)
	return err
}
