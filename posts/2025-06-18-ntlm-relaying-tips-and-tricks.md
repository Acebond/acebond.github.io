---
title: 'NTLM Relaying Tips and Tricks'
---
## NTLM Relay Cheat Sheet
![NTLM Relay](/assets/img/2025-06-18/NTLM Relay.svg)

**Note:** The cheat sheet assumes modern Windows with NTLMv2 being used. NTLMv1 acts the same as HTTP and can be relayed to anything indicated by the “1”. When relaying NTLMv1 using `ntlmrelayx` you must use the `--remove-mic` flag in most cases (such as SMB->LDAP).

## Coerced Authentication

### via RPC Methods (PetitPotam, ShadowCoerce, DFSCoerce, SpoolSample, etc.)
[Coercer](https://github.com/p0dalirius/Coercer) is your best bet. It covers many of the PoCs discovered into a single well tested tool. It also works well using a SOCKS5 proxy. Remember to use the `--auth-type {smb,http}` flag because in many cases you want HTTP.

The Windows WebClient service can be used to determine if coerced auth via WebDAV (HTTP) will work. Any Domain User can determine if that service is running to find good potental relay victums.
- Linux: <https://github.com/Hackndo/WebclientServiceScanner>
- Linux: <https://github.com/Pennyw0rth/NetExec>
- Windows: <https://github.com/MorDavid/SharpWebClientScanner>

Its probably also worth marking all these targets in BloodHound to see which gives paths to DA like so:
#### Neo4j Console Bulk Mark as Owned
```
MATCH (u:Computer) WHERE (u.name IN [
"SERVER1.CONTOSO.COM",
"SERVER2.CONTOSO.COM",
"SERVER3.CONTOSO.COM",
"SERVER4.CONTOSO.COM"
]) SET u.owned = true
```
#### Shortest Path to DA from Owned Computers
The new BloodHound CE uses a different system for owned users, so you'll have to write your own cypher qureies like so:
```
MATCH p=shortestPath((s:Computer)-[:Owns|GenericAll|GenericWrite|WriteOwner|WriteDacl|MemberOf|ForceChangePassword|AllExtendedRights|AddMember|HasSession|GPLink|AllowedToDelegate|CoerceToTGT|AllowedToAct|AdminTo|CanPSRemote|CanRDP|ExecuteDCOM|HasSIDHistory|AddSelf|DCSync|ReadLAPSPassword|ReadGMSAPassword|DumpSMSAPassword|SQLAdmin|AddAllowedToAct|WriteSPN|AddKeyCredentialLink|SyncLAPSPassword|WriteAccountRestrictions|WriteGPLink|GoldenCert|ADCSESC1|ADCSESC3|ADCSESC4|ADCSESC6a|ADCSESC6b|ADCSESC9a|ADCSESC9b|ADCSESC10a|ADCSESC10b|ADCSESC13|SyncedToEntraUser|CoerceAndRelayNTLMToSMB|CoerceAndRelayNTLMToADCS|WriteOwnerLimitedRights|OwnsLimitedRights|CoerceAndRelayNTLMToLDAP|CoerceAndRelayNTLMToLDAPS|Contains|DCFor|SameForestTrust|SpoofSIDHistory|AbuseTGTDelegation*1..]->(t:Group))
WHERE s.enabled = true
AND s.owned = true
AND t.objectid ENDS WITH '-512'
RETURN p
```

Then simply coerce the best targets and compromise them via RBCD or Shadow Creds.



### via Spoofing Attacks (LLMNR, NBT-NS, MDNS, DHCPv6 DNS takeover, ARP, etc.)
- todo

### via Share Poisoning (.lnk, .url, .library-ms, .searchConnector-ms, etc.)
- todo

## Post-Exploitation

### Resource Based Constrained Delegation
- todo

### Shadow Credentials
- todo

## Useful Links
- <https://logan-goins.com/2024-07-23-ldap-relay/>
