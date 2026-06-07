export interface JwtPayload {
  sub: string;
  role: 'ADMIN' | 'USER';
  exp: number;
}
