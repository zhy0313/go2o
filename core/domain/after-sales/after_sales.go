/**
 * Copyright 2015 @ z3q.net.
 * name : after_sales
 * author : jarryliu
 * date : 2016-07-17 11:42
 * description :
 * history :
 */
package afterSales

import (
	"errors"
	"github.com/jsix/gof/db/orm"
	"go2o/core/domain/interface/after-sales"
	"go2o/core/domain/interface/member"
	"go2o/core/domain/interface/order"
	"go2o/core/domain/tmp"
	"strings"
	"time"
)

var _ afterSales.IAfterSalesOrder = new(afterSalesOrderImpl)

type afterSalesOrderImpl struct {
	_value    *afterSales.AfterSalesOrder
	_rep      afterSales.IAfterSalesRep
	_order    order.ISubOrder
	_orderRep order.IOrderRep
}

func NewAfterSalesOrder(v *afterSales.AfterSalesOrder,
	rep afterSales.IAfterSalesRep, orderRep order.IOrderRep,
	memberRep member.IMemberRep) afterSales.IAfterSalesOrder {
	as := newAfterSalesOrder(v, rep, orderRep)
	switch v.Type {
	case afterSales.TypeReturn:
		return newReturnOrderImpl(as, memberRep)
	case afterSales.TypeExchange:
		return newExchangeOrderImpl(as)
	case afterSales.TypeRefund:
		return newRefundOrder(as, memberRep)
	}
	panic(errors.New("不支持的售后单类型"))
}

func newAfterSalesOrder(v *afterSales.AfterSalesOrder,
	rep afterSales.IAfterSalesRep, orderRep order.IOrderRep) *afterSalesOrderImpl {
	return &afterSalesOrderImpl{
		_value:    v,
		_rep:      rep,
		_orderRep: orderRep,
	}
}

// 获取领域编号
func (a *afterSalesOrderImpl) GetDomainId() int {
	return a._value.Id
}

// 获取售后单数据
func (a *afterSalesOrderImpl) Value() afterSales.AfterSalesOrder {
	return *a._value
}

func (a *afterSalesOrderImpl) saveAfterSalesOrder() error {
	if a._value.OrderId <= 0 {
		panic(errors.New("售后单没有绑定订单"))
	}
	if a._value.SnapshotId <= 0 || a._value.Quantity <= 0 {
		panic(errors.New("售后单缺少商品"))
	}
	a._value.UpdateTime = time.Now().Unix()
	id, err := orm.Save(tmp.Db().GetOrm(), a._value, a.GetDomainId())
	if err == nil {
		a._value.Id = id
	}
	return err
}

// 获取订单
func (a *afterSalesOrderImpl) GetOrder() order.ISubOrder {
	if a._order == nil {
		if a._value.OrderId > 0 {
			a._order = a._orderRep.Manager().GetSubOrder(a._value.OrderId)
		}
		if a._order == nil {
			panic(errors.New("售后单对应的订单不存在"))
		}
	}
	return a._order
}

// 设置要退回货物信息
func (a *afterSalesOrderImpl) SetItem(snapshotId int, quantity int) error {
	for _, v := range a.GetOrder().Items() {
		if v.SnapshotId == snapshotId {
			// 判断是否超过数量
			if v.Quantity < quantity {
				return afterSales.ErrOutOfQuantity
			}
			// 设置退回商品
			a._value.SnapshotId = snapshotId
			a._value.Quantity = quantity
			return nil
		}
	}
	return afterSales.ErrNoSuchOrderItem
}

// 提交售后申请
func (a *afterSalesOrderImpl) Submit() (int, error) {
	if a.GetDomainId() > 0 {
		panic(errors.New("售后单已提交"))
	}
	// 售后单未包括商品项
	if a._value.SnapshotId <= 0 || a._value.Quantity <= 0 {
		return 0, afterSales.ErrNoSuchOrderItem
	}
	a._value.Reason = strings.TrimSpace(a._value.Reason)
	if len(a._value.Reason) < 10 {
		return 0, afterSales.ErrReasonLength
	}
	ov := a.GetOrder().GetValue()
	a._value.VendorId = ov.VendorId
	a._value.BuyerId = ov.BuyerId
	a._value.State = afterSales.StatAwaitingVendor
	a._value.CreateTime = time.Now().Unix()
	err := a.saveAfterSalesOrder()
	return a.GetDomainId(), err
}

// 取消申请
func (a *afterSalesOrderImpl) Cancel() error {
	if a._value.State == afterSales.StatCompleted {
		return afterSales.ErrAfterSalesOrderCompleted
	}
	if a._value.State == afterSales.StatCancelled {
		return afterSales.ErrHasCancelled
	}
	a._value.State = afterSales.StatCancelled
	return a.saveAfterSalesOrder()
}

// 同意售后服务
func (a *afterSalesOrderImpl) Agree() error {
	if a._value.State != afterSales.StatAwaitingVendor {
		return afterSales.ErrUnusualStat
	}
	// 判断是否需要审核
	needConfirm := true
	for _, v := range afterSales.IgnoreConfirmTypes {
		if a._value.Type == v {
			needConfirm = false
			break
		}
	}
	// 设置为待审核状态
	a._value.State = afterSales.StatAwaitingConfirm
	// 需要审核
	if needConfirm {
		return a.saveAfterSalesOrder()
	}
	// 不需要审核,直接审核通过
	return a.Confirm()
}

// 拒绝售后服务
func (a *afterSalesOrderImpl) Decline(reason string) error {
	if a._value.State != afterSales.StatAwaitingVendor {
		return afterSales.ErrUnusualStat
	}
	a._value.State = afterSales.StatDeclined
	a._value.VendorRemark = reason
	return a.saveAfterSalesOrder()
}

// 申请调解,只有在商户拒绝后才能申请
func (a *afterSalesOrderImpl) RequestIntercede() error {
	if a._value.State != afterSales.StatDeclined {
		return afterSales.ErrUnusualStat
	}
	a._value.State = afterSales.StatIntercede
	return a.saveAfterSalesOrder()
}

// 系统确认
func (a *afterSalesOrderImpl) Confirm() error {
	if a._value.State == afterSales.StatCompleted {
		return afterSales.ErrAfterSalesOrderCompleted
	}
	if a._value.State == afterSales.StateRejected {
		return afterSales.ErrAfterSalesRejected
	}
	if a._value.State != afterSales.StatAwaitingConfirm {
		return afterSales.ErrUnusualStat
	}
	// 退款,不需要退货,直接进入完成状态
	if a._value.Type == afterSales.TypeRefund {
		return a.awaitingProcess()
	}
	a._value.State = afterSales.StatAwaitingReturnShip
	return a.saveAfterSalesOrder()
}

// 退回售后单
func (a *afterSalesOrderImpl) Reject(remark string) error {
	if a._value.State == afterSales.StatCompleted {
		return afterSales.ErrAfterSalesOrderCompleted
	}
	if a._value.State != afterSales.StatAwaitingConfirm {
		return afterSales.ErrUnusualStat
	}
	a._value.Remark = remark
	a._value.State = afterSales.StateRejected
	return a.saveAfterSalesOrder()
}

// 退回商品
func (a *afterSalesOrderImpl) ReturnShip(spName string, spOrder string, image string) error {
	if a._value.State != afterSales.StatAwaitingReturnShip {
		return afterSales.ErrUnusualStat
	}
	a._value.ReturnSpName = spName
	a._value.ReturnSpOrder = spOrder
	a._value.ReturnSpImage = image
	a._value.State = afterSales.StatReturnShipped
	return a.saveAfterSalesOrder()
}

// 收货, 在商品已退回或尚未发货情况下(线下退货),可以执行此操作
func (a *afterSalesOrderImpl) ReturnReceive() error {
	if a._value.State != afterSales.StatAwaitingReturnShip &&
		a._value.State != afterSales.StatReturnShipped {
		return afterSales.ErrUnusualStat
	}
	return a.awaitingProcess()
}

// 等待处理
func (a *afterSalesOrderImpl) awaitingProcess() error {
	if a._value.State == afterSales.StatCompleted {
		return afterSales.ErrAfterSalesOrderCompleted
	}
	if a._value.State == afterSales.StateRejected {
		return afterSales.ErrAfterSalesRejected
	}

	// 判断状态是否正确
	statOK := a._value.State == afterSales.StatAwaitingReturnShip ||
		a._value.State == afterSales.StatReturnShipped
	if !statOK && a._value.Type == afterSales.TypeRefund {
		statOK = a._value.State == afterSales.StatAwaitingConfirm
	}
	if !statOK {
		return afterSales.ErrNotConfirm
	}

	// 等待处理
	a._value.State = afterSales.StateAwaitingProcess
	return a.saveAfterSalesOrder()
}

// 处理售后单,处理完成后将变为已完成
func (a *afterSalesOrderImpl) Process() error {
	if a._value.State != afterSales.StateAwaitingProcess {
		return afterSales.ErrUnusualStat
	}
	a._value.State = afterSales.StatCompleted
	return a.saveAfterSalesOrder()
}
