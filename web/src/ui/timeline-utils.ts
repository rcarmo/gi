// @ts-nocheck
export const dedupePosts = (items) => {
  const seen = new Set();
  return (items || []).filter((post) => {
    if (!post || seen.has(post.id)) return false;
    seen.add(post.id);
    return true;
  });
};
