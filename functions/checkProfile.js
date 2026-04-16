module.exports = (targetVal, options, context) => {
  // targetVal is the string inside the security array, ( "incomingPaymentCreate")
  
  // Navigate the OpenAPI document in memory to find the catalog
  const profiles = context.document.data?.components?.securitySchemes?.GNAP?.['x-gnap-access-profiles'];

  
  if (!profiles || !profiles[targetVal]) {
    return [{
      message: `Profile '${targetVal}' does not exist in the x-gnap-access-profiles catalog.`
    }];
  }
};