import { PaginatedMeta } from './item.model';

export type OrderStatus =
  | 'PENDING_PAYMENT'
  | 'PAID'
  | 'PROCESSING'
  | 'SHIPPED'
  | 'DELIVERED'
  | 'CANCELLED';

export interface OrderItem {
  id: string;
  itemId: string;
  itemName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

export interface Order {
  id: string;
  userId: string;
  items: OrderItem[];
  total: number;
  status: OrderStatus;
  txHash?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface OrdersResponse {
  orders: Order[];
  meta: PaginatedMeta;
}

export interface ConfirmPaymentRequest {
  txHash: string;
}

export interface UpdateOrderStatusRequest {
  status: 'PROCESSING' | 'SHIPPED' | 'DELIVERED' | 'CANCELLED';
}

export interface OrderFilters {
  page?: number;
  pageSize?: number;
  status?: OrderStatus;
}
