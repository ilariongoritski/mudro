import { useEffect } from 'react'
import { useUploadKeysMutation } from '@/entities/chat/api/chatApi'
import { E2EEService } from '@/entities/chat/model/e2eeService'
import { e2eeStorage } from '@/entities/chat/model/e2eeStorage'
import { useAppSelector } from '@/shared/lib/hooks/storeHooks'

export const useE2EEInitialization = () => {
  const { user, token } = useAppSelector((state) => state.session)
  const [uploadKeys] = useUploadKeysMutation()

  useEffect(() => {
    if (!user || !token) return

    const initKeys = async () => {
      try {
        const existingKeys = await e2eeStorage.getKey(`user_keys_${user.id}`)
        if (existingKeys) {
          console.log('E2EE: Keys already exist for user', user.id)
          return
        }

        console.log('E2EE: Generating new keys for user', user.id)
        const bundle = E2EEService.generateBundle()
        
        // Save to IndexedDB (as Uint8Arrays)
        await e2eeStorage.saveKey(`user_keys_${user.id}`, bundle)

        // Upload to server (as Base64)
        await uploadKeys({
          identity_key: E2EEService.encodeKey(bundle.identity.publicKey),
          signed_prekey: E2EEService.encodeKey(bundle.signedPrekey.publicKey),
          signature: E2EEService.encodeKey(bundle.signature),
          one_time_prekeys: bundle.oneTimePrekeys.map(pk => ({
            id: pk.id,
            key: E2EEService.encodeKey(pk.publicKey)
          }))
        }).unwrap()

        console.log('E2EE: Keys initialized and uploaded successfully')
      } catch (error) {
        console.error('E2EE: Failed to initialize keys', error)
      }
    }

    initKeys()
  }, [user, token, uploadKeys])
}
