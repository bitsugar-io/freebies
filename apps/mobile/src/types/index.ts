export interface Offer {
  offerId: string;
  partnerName: string;
  name: string;
  description: string;
  imageUrl?: string;
}

export interface Rule {
  ruleId: string;
  offerId: string;
  teamId: string;
  condition: Record<string, unknown>;
}

export interface FreebieEvent {
  eventId: string;
  offer: Offer;
  teamId: string;
  teamName: string;
  triggerCondition: string; // Human-readable condition
  regionCode?: string;
  isActive: boolean;
}
