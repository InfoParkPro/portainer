export function isBusinessUpsellHidden() {
  return process.env.PORTAINER_HIDE_BUSINESS_UPSELL === 'true';
}
