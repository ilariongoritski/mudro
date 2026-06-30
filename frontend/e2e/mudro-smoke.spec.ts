import { expect, test } from '@playwright/test'

const authPayload = {
  token: 'e2e-token',
  user: {
    id: 1,
    username: 'e2e_user',
    role: 'user',
  },
}

test.beforeEach(async ({ page }) => {
  await page.route('**/api/auth/me', async (route) => {
    await route.fulfill({ json: authPayload.user })
  })

  await page.route('**/api/auth/refresh', async (route) => {
    await route.fulfill({ json: authPayload })
  })

  await page.route('**/api/auth/login', async (route) => {
    await route.fulfill({ json: authPayload })
  })

  await page.route('**/api/auth/register', async (route) => {
    await route.fulfill({ json: authPayload })
  })

  await page.route('**/api/front**', async (route) => {
    await route.fulfill({
      json: {
        feed: {
          items: [
            {
              id: 1,
              source: 'tg',
              source_post_id: 'e2e-1',
              text: 'E2E smoke post',
              published_at: '2026-06-30T10:00:00Z',
              likes_count: 0,
              comments_count: 0,
              views_count: 1,
              media: [],
              author_name: 'E2E Author',
              source_url: null,
            },
          ],
          next_cursor: null,
        },
        meta: {
          total_posts: 1,
          sources: [{ source: 'tg', posts: 1 }],
        },
      },
    })
  })

  await page.route('**/api/posts/*/like', async (route) => {
    await route.fulfill({ json: { liked: true, likes_count: 1 } })
  })

  await page.route('**/api/posts/*/comments**', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        json: {
          id: 1,
          text: 'E2E comment',
          author_name: 'e2e_user',
          created_at: '2026-06-30T10:01:00Z',
        },
      })
      return
    }
    await route.fulfill({ json: { items: [] } })
  })

  await page.route('**/api/casino/balance', async (route) => {
    await route.fulfill({ json: { balance: 1000, free_spins_balance: 0, currency: 'credits', rtp: 97 } })
  })

  await page.route('**/api/casino/history**', async (route) => {
    await route.fulfill({ json: { items: [] } })
  })

  await page.route('**/api/casino/profile', async (route) => {
    await route.fulfill({
      json: {
        user_id: 1,
        username: 'e2e_user',
        display_name: 'e2e_user',
        balance: 1000,
        total_wagered: 0,
        total_won: 0,
        games_played: 0,
        roulette_rounds_played: 0,
        level: 1,
        xp_progress: 0,
        progress_target: 1000,
        next_level_xp: 1000,
        recent_activity: [],
      },
    })
  })

  await page.route('**/api/casino/activity**', async (route) => {
    await route.fulfill({ json: { items: [] } })
  })

  await page.route('**/api/casino/bonus/state', async (route) => {
    await route.fulfill({ json: { claimed: false, free_spins_balance: 0, recent_claims: [] } })
  })

  await page.route('**/api/casino/roulette/state', async (route) => {
    await route.fulfill({
      json: {
        round_id: 'round-e2e',
        phase: 'idle',
        history: [],
        my_bets: [],
        winning_color: 'unknown',
        display_sequence: [],
        result_sequence: [],
      },
    })
  })

  await page.route('**/api/casino/roulette/history**', async (route) => {
    await route.fulfill({ json: { items: [] } })
  })

  await page.route('**/api/casino/roulette/instant-spin', async (route) => {
    await route.fulfill({
      json: {
        winning_number: 7,
        winning_color: 'red',
        display_sequence: [0, 32, 15, 19, 4, 21, 2, 25, 17, 34, 6, 27, 13, 36, 11, 30, 8, 23, 10, 5, 24, 16, 33, 1, 20, 14, 31, 9, 22, 18, 29, 7],
        result_sequence: [7],
        payout_amount: 20,
        balance: 1020,
        bets: [],
      },
    })
  })

  await page.route('**/api/casino/**', async (route) => {
    await route.fulfill({ json: {} })
  })
})

test('auth pages render and login stores auth state', async ({ page }) => {
  await page.goto('/login', { waitUntil: 'domcontentloaded' })
  await expect(page.getByRole('heading', { name: 'Вход' })).toBeVisible()

  await page.getByLabel('Логин или email').fill('e2e_user')
  await page.getByLabel('Пароль').fill('password123')
  await page.getByRole('button', { name: 'Войти' }).click()

  await expect(page).toHaveURL('/')
  await expect(page.getByRole('heading', { name: 'Лента' })).toBeVisible()
})

test('feed loads with mocked posts', async ({ page }) => {
  await page.goto('/', { waitUntil: 'domcontentloaded' })
  await expect(page.getByText('E2E smoke post')).toBeVisible()
})

test('casino mini app opens roulette tab with mocked balance', async ({ page }) => {
  await page.addInitScript((payload) => {
    window.localStorage.setItem('mudro.session', JSON.stringify(payload))
  }, authPayload)

  await page.goto('/tma/casino', { waitUntil: 'domcontentloaded' })
  await expect(page.getByText('Быстрый игровой режим')).toBeVisible()
  await page.getByRole('button', { name: /Рулетка/ }).click()
  await expect(page.getByText('Рулетка: live')).toBeVisible()
  await expect(page.getByText('Баланс')).toBeVisible()
  await expect(page.getByText('e2e_user · @e2e_user · LVL 1')).toBeVisible()
})
