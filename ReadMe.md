### Cloud Advisor

Cloud Advisor (Optimizer) is a home-region service. This command collects all recommendations for the tenancy using the home region detected from your subscribed regions.

Examples:

  oci-collector cloudadvisor -f json

  oci-collector cloudadvisor -f text

Flags:
- -f, --format: json (default) or text
- -o, --out: optional file path to write the output

So I am viewing this as a fun 'Batman’s utility belt' type of project. What are the common things that we keep sending to our accounts.

# GLOBAL view of my OCI estate (tenancy) is key

Common patterns of finding and reporting on data in OCI.
run as 'user' using config file OR on instance as instance principal
publish as source code or as a binary so that you can take and make your own

This is NOT to replace the CLI or SDK's that are available but has been put together to answer common questions that arise from in our customer and architect community


# Core blocking and tackling
- what are my limits
- Where am I running compute
- Who are all my users
- What are all my policies
- What regions am I subscribed to
- What compartments do i have
- where CAN I launch compute shape x
- what OS is supported by shape
- where do I have ObjectStorage buckets

# Support
- My open support tickets
- My open CAM tickets
- My open 'account' tickets

# What else do?
- Create support ticket?
- what if we add ability to CREATE CAM ticket for limits that are too 'low'?
- Output data to files
- Send output as mail
- Publish as custom metrics


SHOULD the 'core' data (in config) be pushing into a library as almost everything else utilizes those data structures

## Should we refactor to use something like COBRA so adding additional tools is easier?

NOTE: 
if you are using INSTANCE PRINCIPAL you need to have the correct policies in place to allow the instance to read the data (dynamic group and policy etc)



ORACLE AND ITS AFFILIATES DO NOT PROVIDE ANY WARRANTY WHATSOEVER, EXPRESS OR IMPLIED, FOR ANY SOFTWARE, MATERIAL OR CONTENT OF ANY KIND CONTAINED OR PRODUCED WITHIN THIS REPOSITORY, AND IN PARTICULAR SPECIFICALLY DISCLAIM ANY AND ALL IMPLIED WARRANTIES OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY, AND FITNESS FOR A PARTICULAR PURPOSE. FURTHERMORE, ORACLE AND ITS AFFILIATES DO NOT REPRESENT THAT ANY CUSTOMARY SECURITY REVIEW HAS BEEN PERFORMED WITH RESPECT TO ANY SOFTWARE, MATERIAL OR CONTENT CONTAINED OR PRODUCED WITHIN THIS REPOSITORY. IN ADDITION, AND WITHOUT LIMITING THE FOREGOING, THIRD PARTIES MAY HAVE POSTED SOFTWARE, MATERIAL OR CONTENT TO THIS REPOSITORY WITHOUT ANY REVIEW. USE AT YOUR OWN RISK.


### Autonomous Databases

Lists all Autonomous Databases (ATP/ADW/AJD) across your subscribed regions and all compartments.

Examples:

  oci-collector autonomous

This reads credentials from toolkit-config.yaml (configPath/profileName or instance principal) and prints a concise table with region, compartment, name, workload, size and state.