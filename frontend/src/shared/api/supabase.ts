import { createClient } from '@supabase/supabase-js'
import { env } from '@/shared/config/env'

const supabaseUrl = env.supabaseUrl
const supabaseAnonKey = env.supabaseAnonKey

export const supabase = createClient(supabaseUrl, supabaseAnonKey)
