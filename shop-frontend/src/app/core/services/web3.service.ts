import { Injectable, signal } from '@angular/core';
import { BrowserProvider, Contract, parseUnits } from 'ethers';

const SEPOLIA_CHAIN_ID = '0xaa36a7'; // 11155111

const TRANSFER_ABI = [
  'function transfer(address to, uint256 amount) returns (bool)',
];

@Injectable({ providedIn: 'root' })
export class Web3Service {
  connectedAddress = signal<string | null>(null);

  private get ethereum(): any {
    return (window as any)['ethereum'];
  }

  isAvailable(): boolean {
    return !!this.ethereum;
  }

  /** Requests accounts and returns the connected address. */
  async connect(): Promise<string> {
    if (!this.ethereum) throw new Error('MetaMask not installed');
    const provider = new BrowserProvider(this.ethereum);
    const accounts: string[] = await provider.send('eth_requestAccounts', []);
    this.connectedAddress.set(accounts[0]);
    return accounts[0];
  }

  /** Switches MetaMask to Sepolia testnet (adds it if not present). */
  async switchToSepolia(): Promise<void> {
    if (!this.ethereum) throw new Error('MetaMask not installed');
    try {
      await this.ethereum.request({
        method: 'wallet_switchEthereumChain',
        params: [{ chainId: SEPOLIA_CHAIN_ID }],
      });
    } catch (err: any) {
      if (err.code === 4902) {
        await this.ethereum.request({
          method: 'wallet_addEthereumChain',
          params: [{
            chainId: SEPOLIA_CHAIN_ID,
            chainName: 'Sepolia Testnet',
            nativeCurrency: { name: 'ETH', symbol: 'ETH', decimals: 18 },
            rpcUrls: ['https://rpc.sepolia.org'],
            blockExplorerUrls: ['https://sepolia.etherscan.io'],
          }],
        });
      } else {
        throw err;
      }
    }
  }

  /**
   * Transfers mUSDT tokens to the shop wallet.
   * Returns the transaction hash once the tx is mined (1 confirmation).
   */
  async transferUSDT(
    contractAddress: string,
    toAddress: string,
    amountUsdt: number,
  ): Promise<string> {
    if (!this.ethereum) throw new Error('MetaMask not installed');
    const provider = new BrowserProvider(this.ethereum);
    const signer = await provider.getSigner();
    const contract = new Contract(contractAddress, TRANSFER_ABI, signer);
    const amount = parseUnits(amountUsdt.toFixed(18), 18);
    const tx = await (contract['transfer'] as Function)(toAddress, amount);
    await tx.wait(1);
    return tx.hash as string;
  }
}
