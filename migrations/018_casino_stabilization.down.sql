drop trigger if exists tr_audit_casino_accounts on casino_accounts;
drop function if exists audit_casino_accounts();

drop table if exists casino_rtp_tiers;
drop table if exists casino_accounts_audit;
