export interface BlogPost {
  title: string;
  slug: string;
  content: string;
  excerpt?: string | null;
  featured_image?: string | null;
  status: string;
  author_id?: string;
  published_at?: string;
  meta_title?: string | null;
  meta_description?: string | null;
}

export interface LegalDocument {
  type: string;
  title: string;
  content: string;
  effective_date?: string;
  version: string;
}

export interface FAQ {
  question: string;
  answer: string;
  order_index: number;
  is_published: boolean;
}

export interface BlogCategory {
  name: string;
  slug: string;
  description: string;
}

export interface BlogTag {
  name: string;
  slug: string;
}

export interface FeatureCategory {
  name: string;
  icon: string;
  order: number;
  is_active: boolean;
}

export interface FeatureItem {
  category_id: string;
  title: string;
  desc: string;
  icon: string;
  order: number;
  is_active: boolean;
  details_url?: string | null;
}

export interface LeadershipMember {
  name: string;
  role: string;
  image?: string | null;
  bio?: string | null;
  linkedin_url?: string | null;
  twitter_url?: string | null;
  order: number;
  is_active: boolean;
}

export interface DocumentationSection {
  title: string;
  slug: string;
  description?: string;
  order: number;
}

export interface DocumentationArticle {
  section_id: string;
  title: string;
  slug: string;
  content: string;
  order: number;
  is_published: boolean;
  published_at?: string;
}

