package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = "5432"
	user     = "owner"
	password = "1234"
	database = "L0_Database"
)

func PgxConnection() (*sqlx.DB, error) {
	pgxSource := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, database)
	conn, err := sqlx.Connect("postgres", pgxSource)
	if err != nil {
		return nil, err
	}

	return conn, nil

}

func GetAllOrders(conn *sqlx.DB) ([]Order, error) {

	orderUIDs := []string{}

	if err := conn.Select(&orderUIDs, `SELECT "order_uid" FROM orders`); err != nil {
		fmt.Println(err)
		return nil, err
	}

	orders := []Order{}
	for _, uid := range orderUIDs {
		order := Order{}

		rows := conn.QueryRow("select * from orders where order_uid = $1", uid)
		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Delivery.Name,
			&order.Payment.Transaction, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID,
			&order.DateCreated, &order.OofShard,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		rows = conn.QueryRow(`select * from payment where payment_transaction = $1`, order.Payment.Transaction)
		err = rows.Scan(
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
			&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		rows = conn.QueryRow(`select * from delivery where delivery_name = $1`, order.Delivery.Name)
		err = rows.Scan(
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
			&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		items, err := conn.Query("select * from item where order_uid = $1", uid)
		if err != nil {
			return nil, err
		}

		item := Item{}
		for items.Next() {
			err = items.Scan(&item.ChrtID, &order.OrderUID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
			if err != nil {
				return nil, err
			}
			order.Items = append(order.Items, item)
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func InsertOrderToPgx(order Order, conn *sqlx.DB) error {

	delivery := order.Delivery
	_, err := conn.Exec("insert into delivery (delivery_name, phone, zip, city, address, region, email) values($1, $2, $3, $4, $5, $6, $7)",
		delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email)
	if err != nil {
		return err
	}

	payment := order.Payment
	_, err = conn.Exec(
		"insert into payment (payment_transaction, request_id, currency, payment_provider, amount,payment_dt, bank, delivery_cost, goods_total, custom_fee) "+
			"VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount,
		payment.PaymentDt, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee,
	)
	if err != nil {
		return err
	}

	_, err = conn.Exec("insert into orders (order_uid, track_number, entry, delivery, payment, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) "+
		"VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
		order.OrderUID, order.TrackNumber, order.Entry, order.Delivery.Name, order.Payment.Transaction,
		order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey,
		order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		_, err = conn.Exec("insert into item "+
			"(chrt_id, order_uid, track_number, price, rid, item_name, sale, size, total_price, nm_id, brand, status) "+
			"VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
			item.ChrtID, order.OrderUID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size,
			item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
