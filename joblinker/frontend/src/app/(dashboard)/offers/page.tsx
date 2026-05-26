import type { Offer } from '@/shared/types'

import { fetchServer } from '@/lib/api-server'

import { OffersContent } from './OffersContent'

function parseOfferCompensation(offer: Offer): Offer {
  return {
    ...offer,
    compensation: typeof offer.compensation === 'string'
      ? JSON.parse(offer.compensation)
      : offer.compensation,
  }
}

export default async function OffersPage() {
  const data = await fetchServer<{ offers: Offer[] }>('/api/offers')
  const offers = (data?.offers || []).map(parseOfferCompensation)

  return <OffersContent initialOffers={offers} />
}
