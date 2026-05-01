import nacl from 'tweetnacl'
import { E2EEService, type LocalE2EEKeyBundle, type RemoteE2EEKeyBundle } from './e2eeService'
import { e2eeStorage } from './e2eeStorage'

interface E2EESession {
  rootKey: Uint8Array
  sendingChainKey: Uint8Array
  receivingChainKey: Uint8Array
  remoteIdentityKey: Uint8Array
  remoteRatchetKey: Uint8Array
  localRatchetKey: nacl.BoxKeyPair
  ephemeralPublicKey: Uint8Array
}

export class SessionManager {
  static async getSession(userId: string): Promise<E2EESession | null> {
    return (await e2eeStorage.getKey<E2EESession>(`session_${userId}`)) ?? null
  }

  static async createSession(userId: string, aliceKeys: LocalE2EEKeyBundle, bobBundle: RemoteE2EEKeyBundle) {
    const { sharedSecret, ephemeralPublicKey } = E2EEService.establishSession(aliceKeys, bobBundle)
    
    // Initial ratchet step
    const { chainKey, newKeyPair } = E2EEService.ratchetStep(sharedSecret, bobBundle.signedPrekey)

    const session: E2EESession = {
      rootKey: sharedSecret,
      sendingChainKey: chainKey,
      receivingChainKey: chainKey, // For simplicity in v1
      remoteIdentityKey: bobBundle.identity,
      remoteRatchetKey: bobBundle.signedPrekey,
      localRatchetKey: newKeyPair,
      ephemeralPublicKey
    }

    await e2eeStorage.saveKey(`session_${userId}`, session)
    return session
  }

  static async encryptForUser(userId: string, plaintext: string) {
    const session = await this.getSession(userId)
    if (!session) return null

    const { cyphertext, nonce } = E2EEService.encryptMessage(session.sendingChainKey, plaintext)
    
    // Rotate chain key
    session.sendingChainKey = nacl.hash(session.sendingChainKey)
    await e2eeStorage.saveKey(`session_${userId}`, session)

    return { cyphertext, nonce }
  }
}
