import { Item } from './item.model';

export interface CartItem {
  id: string;
  item: Item;
  quantity: number;
  subtotal: number;
}

export interface Cart {
  id: string;
  userId: string;
  items: CartItem[];
  total: number;
  updatedAt: string;
}

export interface AddToCartRequest {
  itemId: string;
  quantity: number;
}

export interface UpdateCartItemRequest {
  quantity: number;
}
