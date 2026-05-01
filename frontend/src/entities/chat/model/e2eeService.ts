import nacl from 'tweetnacl'
import { encodeBase64, decodeBase64 } from 'tweetnacl-util'

export interface LocalE2EEKeyBundle {
  identity: nacl.BoxKeyPair
  signedPrekey: nacl.BoxKeyPair
}

export interface RemoteE2EEKeyBundle {
  identity: Uint8Array
  signedPrekey: Uint8Array
  oneTimePrekey?: Uint8Array
}

export class E2EEService {
  static generateKeyPair() {
    return nacl.box.keyPair()
  }

  static generateBundle() {
    const identityKeyPair = nacl.box.keyPair()
    const signedPrekeyPair = nacl.box.keyPair()
    
    // В реальном приложении мы бы подписывали signedPrekey с помощью identityKey (Ed25519)
    // Но для MVP используем X25519 для обоих и простую заглушку подписи
    const signature = nacl.hash(signedPrekeyPair.publicKey)

    const oneTimePrekeys = Array.from({ length: 10 }).map((_, i) => ({
      id: i,
      ...nacl.box.keyPair(),
    }))

    return {
      identity: identityKeyPair,
      signedPrekey: signedPrekeyPair,
      signature: signature,
      oneTimePrekeys: oneTimePrekeys,
    }
  }

  static encodeKey(key: Uint8Array): string {
    return encodeBase64(key)
  }

  static decodeKey(encoded: string): Uint8Array {
    return decodeBase64(encoded)
  }

  static establishSession(aliceKeys: LocalE2EEKeyBundle, bobBundle: RemoteE2EEKeyBundle) {
    // Alice's ephemeral key
    const ephemeralKeyPair = nacl.box.keyPair()

    // 1. DH1 = DH(AliceIdentity, BobSignedPrekey)
    const dh1 = nacl.box.before(bobBundle.signedPrekey, aliceKeys.identity.secretKey)
    
    // 2. DH2 = DH(AliceEphemeral, BobIdentity)
    const dh2 = nacl.box.before(bobBundle.identity, ephemeralKeyPair.secretKey)

    // 3. DH3 = DH(AliceEphemeral, BobSignedPrekey)
    const dh3 = nacl.box.before(bobBundle.signedPrekey, ephemeralKeyPair.secretKey)

    let sharedSecret = nacl.hash(new Uint8Array([...dh1, ...dh2, ...dh3]))

    // 4. DH4 = DH(AliceEphemeral, BobOneTimePrekey) - Optional
    if (bobBundle.oneTimePrekey) {
      const dh4 = nacl.box.before(bobBundle.oneTimePrekey, ephemeralKeyPair.secretKey)
      sharedSecret = nacl.hash(new Uint8Array([...sharedSecret, ...dh4]))
    }

    return {
      sharedSecret,
      ephemeralPublicKey: ephemeralKeyPair.publicKey
    }
  }

  static ratchetStep(rootKey: Uint8Array, remotePublicKey: Uint8Array) {
    const keyPair = nacl.box.keyPair()
    const dh = nacl.box.before(remotePublicKey, keyPair.secretKey)
    
    // KDF: Используем упрощенный HMAC-SHA512 через nacl.hash
    // В полноценном Signal здесь был бы HKDF
    const salt = rootKey
    const chainKey = nacl.hash(new Uint8Array([...salt, ...dh]))
    
    return {
      chainKey,
      newKeyPair: keyPair
    }
  }

  static encryptMessage(chainKey: Uint8Array, plaintext: string) {
    const messageKey = nacl.hash(chainKey)
    const nonce = nacl.randomBytes(nacl.box.nonceLength)
    const encoded = new TextEncoder().encode(plaintext)
    
    // Нам нужен симметричный шифр, но nacl.box — асимметричный.
    // Трюк: используем nacl.secretbox (XSalsa20-Poly1305) для симметрии.
    const cyphertext = nacl.secretbox(encoded, nonce, messageKey.slice(0, 32))
    
    return {
      cyphertext: encodeBase64(cyphertext),
      nonce: encodeBase64(nonce)
    }
  }

  static decryptMessage(messageKey: Uint8Array, cyphertext: string, nonce: string) {
    const decodedCypher = decodeBase64(cyphertext)
    const decodedNonce = decodeBase64(nonce)
    
    const decrypted = nacl.secretbox.open(decodedCypher, decodedNonce, messageKey.slice(0, 32))
    if (!decrypted) return null
    
    return new TextDecoder().decode(decrypted)
  }
}
