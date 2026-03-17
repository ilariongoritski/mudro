# Public Links

## Verified links
- Self-hosted VPS MVP: `http://91.218.113.247/`
- Public Vercel MVP: `https://frontend-psi-ten-33.vercel.app`

## Verified but not suitable as the main public URL
- Protected Vercel preview:
  `https://frontend-nv1pu0992-goritskimihail-2652s-projects.vercel.app`
  - current behavior: requires auth, returns `401 Unauthorized`

## Recommendation
For the final MVP commit and public sharing:
- keep the self-hosted VPS URL as the runtime baseline
- keep `frontend-psi-ten-33.vercel.app` as the easy external review URL

## Why
- VPS URL is under your direct runtime control
- Vercel public URL is convenient for quick external review and sharing
- the protected preview should not be used as the main public handoff link

