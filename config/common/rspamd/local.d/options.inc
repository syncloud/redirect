// every sender reaching us is an authenticated device, so scanning is about
// catching a compromised one rather than classifying unknown inbound mail
filters = "chartable,dkim,regexp,fuzzy_check";
