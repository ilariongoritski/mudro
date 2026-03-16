alter table if exists post_media_links
  drop constraint if exists post_media_links_post_id_media_asset_id_key;

alter table if exists comment_media_links
  drop constraint if exists comment_media_links_comment_id_media_asset_id_key;
