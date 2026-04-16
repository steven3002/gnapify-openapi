module.exports = (targetVal, options, context) => {
  // targetVal is the entire x-gnap-access-profiles object
  const paths = context.document.data?.paths || {};
  const pathsString = JSON.stringify(paths);
  const errors = [];
  
  // Loop through every profile name defined in the catalog
  for (const profileName of Object.keys(targetVal)) {
    // If the profile name does not appear anywhere in the paths object, flag it
    if (!pathsString.includes(`"${profileName}"`)) {
      errors.push({
        message: `Profile '${profileName}' is declared in the catalog but never referenced by any operation's security array.`,
        path: [profileName] 
      });
    }
  }
  
  return errors;
};