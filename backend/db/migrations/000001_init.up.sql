--
-- PostgreSQL database dump
--

\restrict x2dRdwLoOXbQaHuj4FuoPoUThxvqdiJEfId3pWWIP2BF3RhJYodOpDxksQ178CY

-- Dumped from database version 18.4 (Debian 18.4-1+b1)
-- Dumped by pg_dump version 18.4 (Debian 18.4-1+b1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: softivite; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA softivite;


--
-- Name: thinkce; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA thinkce;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    permissions jsonb DEFAULT '{}'::jsonb
);


--
-- Name: admin_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    user_id uuid NOT NULL,
    admin_role_id uuid NOT NULL
);


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    tenant_id uuid,
    user_id uuid,
    action character varying(50),
    model_name character varying(100),
    object_id character varying(100),
    changes jsonb,
    ip_address character varying(45),
    user_agent text
);


--
-- Name: billing_payments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.billing_payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    subscription_id uuid NOT NULL,
    amount numeric(10,2),
    stripe_invoice_id character varying(100),
    status character varying(20) DEFAULT 'succeeded'::character varying,
    date timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    paystack_reference character varying(100)
);


--
-- Name: billing_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.billing_settings (
    id bigint NOT NULL,
    referral_reward_ghs numeric(10,2) DEFAULT 10
);


--
-- Name: billing_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.billing_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: billing_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.billing_settings_id_seq OWNED BY public.billing_settings.id;


--
-- Name: blog_posts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blog_posts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title character varying(200) NOT NULL,
    slug character varying(200) NOT NULL,
    content text NOT NULL,
    excerpt text,
    featured_image character varying(512),
    status character varying(20) DEFAULT 'draft'::character varying,
    author_id uuid,
    published_at timestamp with time zone,
    meta_title character varying(150),
    meta_description text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: branches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.branches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    unique_id text,
    address text,
    latitude numeric(9,6),
    longitude numeric(9,6),
    phone text,
    primary_color character varying(7) DEFAULT '#4f46e5'::character varying,
    logo character varying(512),
    low_stock_threshold bigint DEFAULT 10,
    currency_symbol character varying(5) DEFAULT 'GH₵'::character varying,
    currency_code character varying(3) DEFAULT 'GHS'::character varying,
    receipt_header text,
    receipt_footer text,
    branch_type character varying(20) DEFAULT 'retail'::character varying,
    last_sync_at timestamp with time zone,
    sync_status character varying(20) DEFAULT 'healthy'::character varying,
    pending_sync_count bigint DEFAULT 0,
    sync_error_message text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: broadcasts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.broadcasts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    title character varying(255) NOT NULL,
    message text NOT NULL,
    type character varying(50) DEFAULT 'info'::character varying,
    created_by uuid,
    target_audience character varying(50) DEFAULT 'all'::character varying
);


--
-- Name: cross_tenant_audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cross_tenant_audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid,
    accessed_tenant_id uuid,
    user_home_tenant_id uuid,
    action_type character varying(20),
    target_model character varying(100),
    target_object_id character varying(100),
    target_object_repr character varying(200),
    description text,
    ip_address character varying(45),
    user_agent text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: domains; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.domains (
    id bigint NOT NULL,
    tenant_id uuid NOT NULL,
    domain character varying(253) NOT NULL,
    is_primary boolean DEFAULT false,
    is_verified boolean DEFAULT false,
    verification_token uuid DEFAULT gen_random_uuid(),
    dns_checked_at timestamp with time zone,
    created_at timestamp with time zone
);


--
-- Name: domains_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.domains_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: domains_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.domains_id_seq OWNED BY public.domains.id;


--
-- Name: goadmin_menu_myid_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goadmin_menu_myid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999999
    CACHE 1;


--
-- Name: goadmin_menu; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_menu (
    id integer DEFAULT nextval('public.goadmin_menu_myid_seq'::regclass) NOT NULL,
    parent_id integer DEFAULT 0 NOT NULL,
    type integer DEFAULT 0,
    "order" integer DEFAULT 0 NOT NULL,
    title character varying(50) NOT NULL,
    header character varying(100),
    plugin_name character varying(100) NOT NULL,
    icon character varying(50) NOT NULL,
    uri character varying(3000) NOT NULL,
    uuid character varying(100),
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_operation_log_myid_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goadmin_operation_log_myid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999999
    CACHE 1;


--
-- Name: goadmin_operation_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_operation_log (
    id integer DEFAULT nextval('public.goadmin_operation_log_myid_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    path character varying(255) NOT NULL,
    method character varying(10) NOT NULL,
    ip character varying(15) NOT NULL,
    input text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_permissions_myid_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goadmin_permissions_myid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999999
    CACHE 1;


--
-- Name: goadmin_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_permissions (
    id integer DEFAULT nextval('public.goadmin_permissions_myid_seq'::regclass) NOT NULL,
    name character varying(50) NOT NULL,
    slug character varying(50) NOT NULL,
    http_method character varying(255),
    http_path text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_role_menu; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_role_menu (
    role_id integer NOT NULL,
    menu_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_role_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_role_permissions (
    role_id integer NOT NULL,
    permission_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_role_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_role_users (
    role_id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_roles_myid_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goadmin_roles_myid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999999
    CACHE 1;


--
-- Name: goadmin_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_roles (
    id integer DEFAULT nextval('public.goadmin_roles_myid_seq'::regclass) NOT NULL,
    name character varying NOT NULL,
    slug character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_session_myid_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goadmin_session_myid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999999
    CACHE 1;


--
-- Name: goadmin_session; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_session (
    id integer DEFAULT nextval('public.goadmin_session_myid_seq'::regclass) NOT NULL,
    sid character varying(50) NOT NULL,
    "values" character varying(3000) NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_site_myid_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goadmin_site_myid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999999
    CACHE 1;


--
-- Name: goadmin_site; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_site (
    id integer DEFAULT nextval('public.goadmin_site_myid_seq'::regclass) NOT NULL,
    key character varying(100) NOT NULL,
    value text NOT NULL,
    type integer DEFAULT 0,
    description character varying(3000),
    state integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_user_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_user_permissions (
    user_id integer NOT NULL,
    permission_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: goadmin_users_myid_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goadmin_users_myid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999999
    CACHE 1;


--
-- Name: goadmin_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goadmin_users (
    id integer DEFAULT nextval('public.goadmin_users_myid_seq'::regclass) NOT NULL,
    username character varying(100) NOT NULL,
    password character varying(100) NOT NULL,
    name character varying(100) NOT NULL,
    avatar character varying(255),
    remember_token character varying(100),
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: ip_allowlists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ip_allowlists (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    ip_address character varying(45) NOT NULL,
    description character varying(255)
);


--
-- Name: legal_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.legal_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    type character varying(50) NOT NULL,
    title character varying(200) NOT NULL,
    content text NOT NULL,
    effective_date timestamp with time zone,
    version character varying(20) DEFAULT '1.0'::character varying
);


--
-- Name: master_api_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.master_api_keys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    key character varying(255) NOT NULL,
    is_active boolean DEFAULT true,
    last_used timestamp with time zone,
    expires_at timestamp with time zone
);


--
-- Name: plan_features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.plan_features (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    plan_id uuid NOT NULL,
    text character varying(255),
    is_available boolean DEFAULT true,
    order_index bigint DEFAULT 0
);


--
-- Name: plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.plans (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(50) NOT NULL,
    description text,
    price numeric(10,2) DEFAULT 0,
    price_ghs numeric(10,2) DEFAULT 0,
    "interval" character varying(20) DEFAULT 'monthly'::character varying,
    trial_days bigint DEFAULT 0,
    stripe_price_id character varying(100),
    max_branches bigint DEFAULT 1,
    max_users bigint DEFAULT 1,
    api_access boolean DEFAULT false,
    api_daily_limit bigint DEFAULT 0,
    is_custom boolean DEFAULT false,
    price_per_branch numeric(10,2) DEFAULT 0,
    price_per_user numeric(10,2) DEFAULT 0,
    price_per_branch_ghs numeric(10,2) DEFAULT 0,
    price_per_user_ghs numeric(10,2) DEFAULT 0,
    features jsonb DEFAULT '{}'::jsonb,
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL,
    paystack_plan_code character varying(100)
);


--
-- Name: pricing_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pricing_plans (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100),
    slug character varying(50),
    price_monthly numeric(10,2),
    price_yearly numeric(10,2),
    currency character varying(3) DEFAULT 'USD'::character varying,
    description text,
    is_popular boolean DEFAULT false,
    button_text character varying(50) DEFAULT 'Get Started'::character varying,
    order_index bigint DEFAULT 0,
    max_branches bigint DEFAULT 1,
    max_staff bigint DEFAULT 1
);


--
-- Name: promo_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promo_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    code character varying(50) NOT NULL,
    discount_type character varying(20) DEFAULT 'percentage'::character varying,
    discount_value numeric(10,2),
    max_uses bigint DEFAULT 0,
    current_uses bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    valid_from timestamp with time zone,
    valid_until timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: referral_rewards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.referral_rewards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    referrer_id uuid NOT NULL,
    referred_tenant_id uuid NOT NULL,
    reward_amount numeric(10,2),
    is_applied boolean DEFAULT false,
    applied_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: seo_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.seo_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    meta_title character varying(150),
    meta_description text,
    keywords character varying(255),
    og_title character varying(150),
    og_description text,
    og_image character varying(512),
    google_analytics_id character varying(50),
    facebook_pixel_id character varying(50),
    homepage_video_id character varying(50) DEFAULT 'dQw4w9WgXcQ'::character varying,
    contact_email character varying(254),
    support_email character varying(254),
    contact_phone character varying(50),
    contact_address text,
    office_hours character varying(100),
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    plan_id uuid,
    stripe_subscription_id character varying(100),
    stripe_customer_id character varying(100),
    status character varying(20) DEFAULT 'trialing'::character varying,
    current_period_end timestamp with time zone,
    cancel_at_period_end boolean DEFAULT false,
    api_requests_today bigint DEFAULT 0,
    api_requests_month bigint DEFAULT 0,
    api_last_reset_date timestamp with time zone,
    api_month_reset_date timestamp with time zone,
    custom_branches_count bigint,
    custom_users_count bigint,
    version bigint DEFAULT 1 NOT NULL,
    paystack_subscription_code character varying(100),
    paystack_customer_code character varying(100),
    last_billing_email_at timestamp with time zone
);


--
-- Name: tenant_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_metrics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    total_products bigint DEFAULT 0,
    total_orders bigint DEFAULT 0,
    total_customers bigint DEFAULT 0,
    total_branches bigint DEFAULT 0,
    total_revenue numeric(15,2) DEFAULT 0,
    total_customer_debt numeric(15,2) DEFAULT 0,
    last_updated timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: tenants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    subdomain character varying(100) NOT NULL,
    schema_name character varying(100),
    tenant_type character varying(20) DEFAULT 'standard'::character varying,
    logo character varying(512),
    address text,
    pos_api_key text,
    is_sandbox boolean DEFAULT false,
    sandbox_wipe_at timestamp with time zone,
    has_used_trial boolean DEFAULT false,
    referral_code character varying(20),
    referred_by_id uuid,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_on timestamp with time zone,
    status character varying(20) DEFAULT 'active'::character varying
);


--
-- Name: user_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_profiles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid,
    role character varying(50) DEFAULT 'sales'::character varying,
    can_perform_credit_sales boolean DEFAULT false,
    base_salary numeric(12,2) DEFAULT 0,
    hourly_rate numeric(10,2) DEFAULT 0,
    payment_method character varying(20) DEFAULT 'cash'::character varying,
    bank_details jsonb DEFAULT '{}'::jsonb,
    is2_fa_enabled boolean DEFAULT false,
    otp_secret text,
    pos_pin text,
    is_email_verified boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    is_2fa_enabled boolean DEFAULT false
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username character varying(150) NOT NULL,
    email character varying(254) NOT NULL,
    password character varying(255) NOT NULL,
    first_name character varying(150),
    last_name character varying(150),
    is_active boolean DEFAULT true,
    is_superuser boolean DEFAULT false,
    is_staff boolean DEFAULT false,
    last_login timestamp with time zone,
    date_joined timestamp with time zone,
    token_version bigint DEFAULT 1,
    deleted_at timestamp with time zone,
    reset_token character varying(100),
    reset_token_expiry timestamp with time zone,
    phone character varying(20)
);


--
-- Name: abandoned_carts; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.abandoned_carts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    email character varying(255) NOT NULL,
    cart_data jsonb,
    is_recovered boolean DEFAULT false,
    email_sent boolean DEFAULT false
);


--
-- Name: api_keys; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.api_keys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    name character varying(100),
    key_prefix character varying(8),
    key_hash character varying(128),
    is_active boolean DEFAULT true,
    is_sandbox boolean DEFAULT false,
    last_used_at timestamp with time zone
);


--
-- Name: api_request_logs; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.api_request_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    tenant_id uuid,
    user_id uuid,
    method character varying(10),
    endpoint character varying(255),
    status_code bigint,
    response_time_ms bigint,
    ip_address character varying(45),
    user_agent text,
    request_body jsonb,
    response_body jsonb
);


--
-- Name: appointments; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.appointments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    customer_id uuid,
    customer_name character varying(200),
    customer_phone character varying(30),
    customer_email character varying(254),
    service_id uuid NOT NULL,
    staff_member_id uuid,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    status character varying(20) DEFAULT 'scheduled'::character varying,
    notes text,
    order_id uuid
);


--
-- Name: attendances; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.attendances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    staff_id uuid NOT NULL,
    clock_in timestamp with time zone,
    clock_out timestamp with time zone,
    metadata jsonb,
    status character varying(20) DEFAULT 'present'::character varying
);


--
-- Name: audit_logs; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    tenant_id uuid,
    user_id uuid,
    action character varying(50),
    model_name character varying(100),
    object_id character varying(100),
    changes jsonb,
    ip_address character varying(45),
    user_agent text
);


--
-- Name: branches; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.branches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    unique_id text,
    address text,
    latitude numeric(9,6),
    longitude numeric(9,6),
    phone text,
    primary_color character varying(7) DEFAULT '#4f46e5'::character varying,
    logo character varying(512),
    low_stock_threshold bigint DEFAULT 10,
    currency_symbol character varying(5) DEFAULT 'GH₵'::character varying,
    currency_code character varying(3) DEFAULT 'GHS'::character varying,
    receipt_header text,
    receipt_footer text,
    branch_type character varying(20) DEFAULT 'retail'::character varying,
    last_sync_at timestamp with time zone,
    sync_status character varying(20) DEFAULT 'healthy'::character varying,
    pending_sync_count bigint DEFAULT 0,
    sync_error_message text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: cash_drawer_sessions; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.cash_drawer_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid NOT NULL,
    user_id uuid NOT NULL,
    opening_balance numeric(12,2) NOT NULL,
    closing_balance numeric(12,2),
    opened_at timestamp with time zone,
    closed_at timestamp with time zone,
    notes text
);


--
-- Name: categories; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    image character varying(255),
    color character varying(20) DEFAULT 'blue'::character varying
);


--
-- Name: commission_rules; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.commission_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    name character varying(100) NOT NULL,
    min_sales_amount numeric(12,2) DEFAULT 0,
    commission_percentage numeric(5,2) DEFAULT 0,
    flat_bonus numeric(10,2) DEFAULT 0,
    is_active boolean DEFAULT true
);


--
-- Name: coupons; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.coupons (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    code character varying(50) NOT NULL,
    discount_type character varying(20) NOT NULL,
    value numeric(10,2) NOT NULL,
    min_purchase numeric(10,2) DEFAULT 0,
    is_active boolean DEFAULT true,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    usage_limit bigint DEFAULT 100,
    used_count bigint DEFAULT 0
);


--
-- Name: crm_settings; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.crm_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    points_per_currency numeric(5,2) DEFAULT 1,
    redemption_rate numeric(5,2) DEFAULT 0.01,
    monthly_sales_goal numeric(12,2) DEFAULT 50000
);


--
-- Name: customer_feedbacks; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.customer_feedbacks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    customer_id uuid NOT NULL,
    order_id uuid,
    rating bigint DEFAULT 5,
    comment text,
    is_public boolean DEFAULT false
);


--
-- Name: customer_segments; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.customer_segments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    criteria_json jsonb
);


--
-- Name: customer_tiers; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.customer_tiers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(50) NOT NULL,
    min_spend numeric(12,2) DEFAULT 0,
    discount_percentage numeric(5,2) DEFAULT 0,
    color character varying(20) DEFAULT 'blue'::character varying,
    icon character varying(50) DEFAULT 'star'::character varying
);


--
-- Name: customers; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.customers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(200) NOT NULL,
    phone text,
    email text,
    address text,
    password_hash character varying(255),
    is_registered boolean DEFAULT false,
    tier_id uuid,
    total_spend numeric(12,2) DEFAULT 0,
    order_count bigint DEFAULT 0,
    loyalty_pts numeric(10,2) DEFAULT 0,
    store_credit numeric(12,2) DEFAULT 0,
    debt_balance numeric(12,2) DEFAULT 0,
    accepts_marketing boolean DEFAULT true,
    last_visit timestamp with time zone,
    date_of_birth timestamp with time zone,
    notes text,
    customer_type character varying(20) DEFAULT 'retail'::character varying
);


--
-- Name: delivery_drivers; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.delivery_drivers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    phone character varying(20) NOT NULL,
    vehicle_info character varying(100),
    current_status character varying(20) DEFAULT 'available'::character varying,
    lat numeric(10,8),
    lng numeric(11,8)
);


--
-- Name: delivery_orders; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.delivery_orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    order_id uuid NOT NULL,
    driver_id uuid,
    status character varying(20) DEFAULT 'pending'::character varying,
    tracking_link character varying(255),
    delivery_notes text,
    delivery_fee numeric(12,2) DEFAULT 0
);


--
-- Name: dining_tables; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.dining_tables (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    name character varying(50) NOT NULL,
    capacity bigint DEFAULT 4,
    status character varying(20) DEFAULT 'available'::character varying,
    qr_code_url character varying(512),
    position_x bigint DEFAULT 0,
    position_y bigint DEFAULT 0,
    is_active boolean DEFAULT true
);


--
-- Name: discount_codes; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.discount_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    type text NOT NULL,
    value numeric NOT NULL,
    status text DEFAULT 'active'::text,
    max_uses bigint,
    current_uses bigint DEFAULT 0,
    valid_from timestamp with time zone,
    valid_until timestamp with time zone,
    min_order_value numeric,
    points_required bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: domains; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.domains (
    id bigint NOT NULL,
    tenant_id uuid NOT NULL,
    domain character varying(253) NOT NULL,
    is_primary boolean DEFAULT false,
    is_verified boolean DEFAULT false,
    verification_token uuid DEFAULT gen_random_uuid(),
    dns_checked_at timestamp with time zone,
    created_at timestamp with time zone
);


--
-- Name: domains_id_seq; Type: SEQUENCE; Schema: softivite; Owner: -
--

CREATE SEQUENCE softivite.domains_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: domains_id_seq; Type: SEQUENCE OWNED BY; Schema: softivite; Owner: -
--

ALTER SEQUENCE softivite.domains_id_seq OWNED BY softivite.domains.id;


--
-- Name: expense_categories; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.expense_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(20) DEFAULT 'variable'::character varying,
    description text,
    monthly_budget numeric(12,2) DEFAULT 0
);


--
-- Name: expenses; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.expenses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    category_id uuid NOT NULL,
    amount numeric(12,2) NOT NULL,
    date timestamp with time zone,
    description text,
    receipt_url character varying(512),
    is_recurring boolean DEFAULT false,
    recurrence_interval character varying(20) DEFAULT ''::character varying,
    created_by_id uuid
);


--
-- Name: external_systems; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.external_systems (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    developer_id uuid NOT NULL,
    name character varying(150) NOT NULL,
    description text,
    client_id uuid DEFAULT gen_random_uuid(),
    client_secret_hash character varying(128),
    redirect_uris jsonb,
    webhook_url character varying(512),
    icon character varying(50) DEFAULT 'rocket_launch'::character varying,
    is_public boolean DEFAULT false,
    is_active boolean DEFAULT true
);


--
-- Name: gift_cards; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.gift_cards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    code character varying(50) NOT NULL,
    initial_balance numeric(10,2) NOT NULL,
    current_balance numeric(10,2) NOT NULL,
    purchaser_id uuid,
    recipient_email character varying(254),
    expires_at timestamp with time zone,
    is_active boolean DEFAULT true,
    status character varying(20) DEFAULT 'active'::character varying
);


--
-- Name: honeypot_attempts; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.honeypot_attempts (
    id bigint NOT NULL,
    username character varying(255),
    password character varying(255),
    ip_address character varying(45),
    user_agent text,
    path character varying(255) DEFAULT '/admin/'::character varying,
    "timestamp" timestamp with time zone
);


--
-- Name: honeypot_attempts_id_seq; Type: SEQUENCE; Schema: softivite; Owner: -
--

CREATE SEQUENCE softivite.honeypot_attempts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: honeypot_attempts_id_seq; Type: SEQUENCE OWNED BY; Schema: softivite; Owner: -
--

ALTER SEQUENCE softivite.honeypot_attempts_id_seq OWNED BY softivite.honeypot_attempts.id;


--
-- Name: journal_entries; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.journal_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    reference_id uuid,
    reference_type character varying(50),
    description text
);


--
-- Name: kds_tickets; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.kds_tickets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    order_id uuid NOT NULL,
    table_id uuid,
    status character varying(20) DEFAULT 'pending'::character varying,
    kitchen_notes text,
    is_rush boolean DEFAULT false,
    started_at timestamp with time zone,
    completed_at timestamp with time zone
);


--
-- Name: leave_requests; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.leave_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    staff_id uuid NOT NULL,
    leave_type character varying(20) NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    reason text,
    status character varying(20) DEFAULT 'pending'::character varying,
    reviewed_by_id uuid,
    reviewed_at timestamp with time zone
);


--
-- Name: ledger_accounts; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.ledger_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(50) NOT NULL,
    code character varying(20),
    description text
);


--
-- Name: ledger_lines; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.ledger_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    journal_entry_id uuid NOT NULL,
    account_id uuid NOT NULL,
    amount numeric(12,2) NOT NULL,
    is_debit boolean NOT NULL
);


--
-- Name: loyalty_transactions; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.loyalty_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    tenant_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    order_id uuid,
    points numeric(10,2) NOT NULL,
    transaction_type character varying(20) NOT NULL,
    description text
);


--
-- Name: marketing_campaigns; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.marketing_campaigns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    campaign_type character varying(10) DEFAULT 'email'::character varying,
    subject character varying(255),
    message text NOT NULL,
    coupon_code character varying(50),
    status character varying(20) DEFAULT 'draft'::character varying,
    target_tier_id uuid,
    segment_id uuid,
    is_automated boolean DEFAULT false,
    trigger_event character varying(30) DEFAULT 'manual'::character varying,
    open_count bigint DEFAULT 0,
    click_count bigint DEFAULT 0,
    conversion_count bigint DEFAULT 0,
    revenue_generated numeric(12,2) DEFAULT 0,
    scheduled_at timestamp with time zone,
    sent_at timestamp with time zone,
    last_run_at timestamp with time zone
);


--
-- Name: newsletter_subscriptions; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.newsletter_subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    email character varying(255) NOT NULL,
    is_active boolean DEFAULT true
);


--
-- Name: notification_settings; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.notification_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    user_id uuid NOT NULL,
    email_notifications boolean DEFAULT true,
    low_stock_alerts boolean DEFAULT true,
    sales_reports boolean DEFAULT true,
    security_alerts boolean DEFAULT true,
    system_alerts boolean DEFAULT true
);


--
-- Name: notifications; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    user_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    message text NOT NULL,
    link character varying(255),
    is_read boolean DEFAULT false,
    notification_type character varying(20) DEFAULT 'info'::character varying,
    category character varying(20) DEFAULT 'general'::character varying
);


--
-- Name: order_items; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    order_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    quantity numeric(10,3) NOT NULL,
    unit_price numeric(12,2) NOT NULL,
    discount numeric(12,2) DEFAULT 0,
    total numeric(12,2) NOT NULL,
    cost_price numeric(12,2) DEFAULT 0
);


--
-- Name: orders; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1,
    branch_id uuid,
    order_number character varying(50) NOT NULL,
    customer_id uuid,
    cashier_id uuid,
    subtotal numeric(12,2) DEFAULT 0,
    tax numeric(12,2) DEFAULT 0,
    discount numeric(12,2) DEFAULT 0,
    total numeric(12,2) NOT NULL,
    amount_paid numeric(12,2) DEFAULT 0,
    status character varying(20) DEFAULT 'completed'::character varying,
    payment_status character varying(20) DEFAULT 'paid'::character varying,
    payment_method character varying(20),
    order_type character varying(20) DEFAULT 'in_store'::character varying,
    notes text,
    receipt_token character varying(64)
);


--
-- Name: payment_methods; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.payment_methods (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    provider character varying(20) NOT NULL,
    is_active boolean DEFAULT true,
    api_key_hint character varying(50)
);


--
-- Name: payments; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    order_id uuid NOT NULL,
    payment_method_id uuid NOT NULL,
    amount numeric(12,2) NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    transaction_id character varying(255),
    metadata jsonb,
    error_message text
);


--
-- Name: payroll_periods; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.payroll_periods (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    is_processed boolean DEFAULT false,
    processed_at timestamp with time zone
);


--
-- Name: payroll_records; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.payroll_records (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    period_id uuid NOT NULL,
    staff_id uuid NOT NULL,
    base_salary_snapshot numeric(12,2) NOT NULL,
    total_commission numeric(12,2) DEFAULT 0,
    bonus numeric(12,2) DEFAULT 0,
    deductions numeric(12,2) DEFAULT 0,
    net_pay numeric(12,2) NOT NULL,
    is_paid boolean DEFAULT false,
    paid_at timestamp with time zone,
    payment_reference character varying(100)
);


--
-- Name: print_jobs; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.print_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid NOT NULL,
    document_type character varying(50) NOT NULL,
    content text NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    printed_at timestamp with time zone
);


--
-- Name: product_components; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.product_components (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    composite_product_id uuid NOT NULL,
    component_product_id uuid NOT NULL,
    quantity numeric(10,4) NOT NULL
);


--
-- Name: product_histories; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.product_histories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    product_id uuid NOT NULL,
    user_id uuid,
    field character varying(50),
    old_value text,
    new_value text
);


--
-- Name: product_image_galleries; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.product_image_galleries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    product_id uuid NOT NULL,
    image_url character varying(512) NOT NULL,
    alt_text character varying(255),
    "order" bigint DEFAULT 0
);


--
-- Name: product_reviews; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.product_reviews (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    product_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    rating bigint NOT NULL,
    comment text,
    is_visible boolean DEFAULT true
);


--
-- Name: product_variants; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.product_variants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    product_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    sku character varying(100) NOT NULL,
    barcode character varying(100),
    price_override numeric(12,2),
    cost_override numeric(12,2),
    current_stock numeric(10,4) DEFAULT 0,
    attributes jsonb,
    image character varying(255),
    CONSTRAINT chk_product_variants_attributes CHECK ((jsonb_typeof(attributes) = 'object'::text))
);


--
-- Name: products; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1,
    branch_id uuid,
    name character varying(200) NOT NULL,
    description text,
    sku character varying(100) NOT NULL,
    barcode character varying(100),
    category_id uuid,
    cost_price numeric(12,2) DEFAULT 0,
    selling_price numeric(12,2) NOT NULL,
    wholesale_price numeric(12,2) DEFAULT 0,
    track_inventory boolean DEFAULT true,
    current_stock numeric(10,4) DEFAULT 0,
    reorder_level numeric(10,4) DEFAULT 0,
    stock_unit character varying(50) DEFAULT 'pcs'::character varying,
    has_variants boolean DEFAULT false,
    is_composite boolean DEFAULT false,
    is_active boolean DEFAULT true,
    is_online boolean DEFAULT true,
    image character varying(255),
    supplier_id uuid,
    last_received_date timestamp with time zone,
    expiry_date timestamp with time zone,
    manufacturing_date timestamp with time zone,
    minimum_wholesale_quantity numeric(10,4) DEFAULT 1,
    batch_number character varying(100),
    invoice_waybill_number character varying(100),
    country_of_origin character varying(100),
    manufacturer_name character varying(200),
    manufacturer_address text
);


--
-- Name: promotions; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.promotions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'draft'::text,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    description text,
    points_required bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: purchase_order_items; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.purchase_order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    po_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    quantity_ordered numeric(10,4) NOT NULL,
    quantity_received numeric(10,4) DEFAULT 0,
    unit_cost numeric(12,2) NOT NULL
);


--
-- Name: purchase_orders; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.purchase_orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    po_number character varying(50) NOT NULL,
    supplier_id uuid NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying,
    total_amount numeric(12,2) DEFAULT 0,
    expected_date timestamp with time zone,
    notes text
);


--
-- Name: quotation_items; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.quotation_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    quotation_id uuid NOT NULL,
    product_id uuid NOT NULL,
    quantity bigint DEFAULT 1,
    unit_price numeric(12,2) NOT NULL,
    discount numeric(12,2) DEFAULT 0,
    total_price numeric(12,2) NOT NULL
);


--
-- Name: quotations; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.quotations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    customer_id uuid NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying,
    quote_number character varying(30) NOT NULL,
    subtotal numeric(12,2) DEFAULT 0,
    tax_amount numeric(12,2) DEFAULT 0,
    total_amount numeric(12,2) DEFAULT 0,
    notes text,
    internal_notes text,
    valid_until timestamp with time zone,
    created_by_id uuid,
    reviewed_by_id uuid,
    converted_order_id uuid
);


--
-- Name: return_items; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.return_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    return_id uuid NOT NULL,
    product_id uuid,
    quantity numeric(10,3) DEFAULT 1,
    condition character varying(20) DEFAULT 'opened'::character varying,
    restock boolean DEFAULT false,
    unit_price numeric(12,2) NOT NULL
);


--
-- Name: returns; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.returns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    order_id uuid NOT NULL,
    customer_id uuid,
    reason character varying(50) NOT NULL,
    reason_detail text,
    status character varying(20) DEFAULT 'pending'::character varying,
    refund_method character varying(20) DEFAULT 'original'::character varying,
    refund_amount numeric(12,2) DEFAULT 0,
    restocking_fee numeric(12,2) DEFAULT 0,
    created_by_id uuid,
    approved_by_id uuid,
    approved_at timestamp with time zone,
    completed_at timestamp with time zone
);


--
-- Name: service_categories; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.service_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    icon character varying(50) DEFAULT 'spa'::character varying,
    color character varying(20) DEFAULT 'purple'::character varying
);


--
-- Name: service_commission_records; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.service_commission_records (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    staff_member_id uuid NOT NULL,
    rule_id uuid,
    order_id uuid NOT NULL,
    amount numeric(10,2) NOT NULL,
    is_paid boolean DEFAULT false,
    paid_at timestamp with time zone,
    notes text
);


--
-- Name: service_commission_rules; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.service_commission_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    staff_member_id uuid NOT NULL,
    commission_type character varying(20) DEFAULT 'percentage'::character varying,
    value numeric(8,2) NOT NULL,
    applies_to character varying(20) DEFAULT 'all_sales'::character varying,
    is_active boolean DEFAULT true
);


--
-- Name: services; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.services (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    category_id uuid,
    name character varying(200) NOT NULL,
    description text,
    duration_minutes bigint DEFAULT 30,
    price numeric(10,2) NOT NULL,
    default_staff_id uuid,
    image character varying(512),
    is_active boolean DEFAULT true
);


--
-- Name: shift_swap_requests; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.shift_swap_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    requesting_staff_id uuid NOT NULL,
    target_staff_id uuid,
    original_shift_id uuid NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    notes text
);


--
-- Name: shifts; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.shifts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid NOT NULL,
    user_id uuid NOT NULL,
    start_time timestamp with time zone,
    end_time timestamp with time zone
);


--
-- Name: split_bill_groups; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.split_bill_groups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    table_id uuid,
    original_order_id uuid,
    notes text
);


--
-- Name: staff_achievements; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.staff_achievements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    staff_id uuid NOT NULL,
    badge_name character varying(100) NOT NULL,
    badge_icon character varying(50) DEFAULT 'stars'::character varying,
    description text
);


--
-- Name: stock_batches; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.stock_batches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid NOT NULL,
    product_id uuid NOT NULL,
    batch_number character varying(100) NOT NULL,
    quantity numeric(10,4) NOT NULL,
    expiry_date timestamp with time zone,
    manufacture_date timestamp with time zone
);


--
-- Name: stock_movements; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.stock_movements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    quantity numeric(10,4) NOT NULL,
    previous_stock numeric(10,4) NOT NULL,
    new_stock numeric(10,4) NOT NULL,
    reason character varying(50) NOT NULL,
    reference_id character varying(100),
    user_id uuid
);


--
-- Name: stock_transfer_items; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.stock_transfer_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    transfer_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    quantity numeric(10,4) NOT NULL
);


--
-- Name: stock_transfers; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.stock_transfers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    reference_no character varying(50) NOT NULL,
    from_branch_id uuid NOT NULL,
    to_branch_id uuid NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    notes text,
    created_by_id uuid,
    shipped_at timestamp with time zone,
    received_at timestamp with time zone
);


--
-- Name: stocktake_entries; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.stocktake_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    session_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    expected_stock numeric(10,4) NOT NULL,
    actual_stock numeric(10,4) NOT NULL,
    difference numeric(10,4) NOT NULL
);


--
-- Name: stocktake_sessions; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.stocktake_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid,
    name character varying(100) NOT NULL,
    status character varying(20) DEFAULT 'in_progress'::character varying,
    notes text,
    access_token uuid,
    created_by_id uuid,
    completed_at timestamp with time zone
);


--
-- Name: storefront_settings; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.storefront_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    default_branch_id uuid,
    is_active boolean DEFAULT false,
    slug character varying(100),
    store_view_type character varying(20) DEFAULT 'branch'::character varying,
    store_name character varying(100),
    banner_image character varying(255),
    logo_image character varying(255),
    primary_color character varying(7) DEFAULT '#3b82f6'::character varying,
    welcome_message text,
    about_text text,
    allow_pickup boolean DEFAULT true,
    allow_delivery boolean DEFAULT false,
    delivery_fee numeric(8,2) DEFAULT 0,
    min_order_amount numeric(10,2) DEFAULT 0,
    enable_stripe boolean DEFAULT false,
    enable_paystack boolean DEFAULT false,
    paystack_public_key character varying(255),
    paystack_subaccount_code character varying(100),
    enable_mobile_money boolean DEFAULT false,
    flash_sale_end_time timestamp with time zone
);


--
-- Name: supplier_ledger_entries; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.supplier_ledger_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    supplier_id uuid NOT NULL,
    entry_type character varying(50) NOT NULL,
    reference_id character varying(100),
    amount numeric(12,2) NOT NULL,
    balance numeric(12,2) NOT NULL,
    notes text
);


--
-- Name: supplier_products; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.supplier_products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    supplier_id uuid NOT NULL,
    product_id uuid NOT NULL,
    supplier_sku character varying(100),
    unit_cost numeric(12,2) DEFAULT 0 NOT NULL
);


--
-- Name: supplier_profiles; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.supplier_profiles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    user_id uuid NOT NULL,
    supplier_id uuid NOT NULL,
    role character varying(50) DEFAULT 'contact'::character varying,
    is_active boolean DEFAULT true
);


--
-- Name: suppliers; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.suppliers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(200) NOT NULL,
    contact_person character varying(150),
    email text,
    phone text,
    address text,
    tax_number character varying(50),
    payment_terms text,
    notes text,
    credit_balance numeric(12,2) DEFAULT 0,
    is_active boolean DEFAULT true
);


--
-- Name: tax_configurations; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.tax_configurations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    tax_type character varying(20) DEFAULT 'sales_tax'::character varying,
    tax_rate numeric(5,2) DEFAULT 0,
    tax_number character varying(50),
    include_tax_in_prices boolean DEFAULT false,
    is_active boolean DEFAULT true
);


--
-- Name: tenant_integrations; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.tenant_integrations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    provider character varying(50) NOT NULL,
    access_token text,
    refresh_token text,
    expires_at timestamp with time zone,
    config jsonb,
    is_active boolean DEFAULT true,
    last_sync_at timestamp with time zone
);


--
-- Name: webhook_endpoints; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.webhook_endpoints (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    external_system_id uuid,
    url character varying(512) NOT NULL,
    secret character varying(64) NOT NULL,
    is_active boolean DEFAULT true,
    events jsonb
);


--
-- Name: webhook_events; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.webhook_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    endpoint_id uuid NOT NULL,
    event_type character varying(50) NOT NULL,
    payload jsonb,
    signature character varying(128),
    status_code bigint,
    response_body text,
    error_message text
);


--
-- Name: wishlists; Type: TABLE; Schema: softivite; Owner: -
--

CREATE TABLE softivite.wishlists (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    customer_id uuid NOT NULL,
    product_id uuid NOT NULL
);


--
-- Name: abandoned_carts; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.abandoned_carts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email character varying(255) NOT NULL,
    cart_data jsonb,
    is_recovered boolean DEFAULT false,
    email_sent boolean DEFAULT false,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: api_keys; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.api_keys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    name character varying(100),
    key_prefix character varying(8),
    key_hash character varying(128),
    is_active boolean DEFAULT true,
    is_sandbox boolean DEFAULT false,
    last_used_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: api_request_logs; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.api_request_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid,
    user_id uuid,
    method character varying(10),
    endpoint character varying(255),
    status_code bigint,
    response_time_ms bigint,
    ip_address character varying(45),
    user_agent text,
    request_body jsonb,
    response_body jsonb,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: appointments; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.appointments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    customer_id uuid,
    customer_name character varying(200),
    customer_phone character varying(30),
    customer_email character varying(254),
    service_id uuid NOT NULL,
    staff_member_id uuid,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    status character varying(20) DEFAULT 'scheduled'::character varying,
    notes text,
    order_id uuid,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: attendances; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.attendances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    staff_id uuid NOT NULL,
    clock_in timestamp with time zone,
    clock_out timestamp with time zone,
    metadata jsonb,
    status character varying(20) DEFAULT 'present'::character varying,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: audit_logs; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid,
    user_id uuid,
    action character varying(50),
    model_name character varying(100),
    object_id character varying(100),
    changes jsonb,
    ip_address character varying(45),
    user_agent text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: branches; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.branches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    unique_id text,
    address text,
    latitude numeric(9,6),
    longitude numeric(9,6),
    phone text,
    primary_color character varying(7) DEFAULT '#4f46e5'::character varying,
    logo character varying(512),
    low_stock_threshold bigint DEFAULT 10,
    currency_symbol character varying(5) DEFAULT 'GH₵'::character varying,
    currency_code character varying(3) DEFAULT 'GHS'::character varying,
    receipt_header text,
    receipt_footer text,
    branch_type character varying(20) DEFAULT 'retail'::character varying,
    last_sync_at timestamp with time zone,
    sync_status character varying(20) DEFAULT 'healthy'::character varying,
    pending_sync_count bigint DEFAULT 0,
    sync_error_message text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: cash_drawer_sessions; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.cash_drawer_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid NOT NULL,
    user_id uuid NOT NULL,
    opening_balance numeric(12,2) NOT NULL,
    closing_balance numeric(12,2),
    opened_at timestamp with time zone,
    closed_at timestamp with time zone,
    notes text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: categories; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(100) NOT NULL,
    description text,
    image character varying(255),
    color character varying(20) DEFAULT 'blue'::character varying,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: coupons; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.coupons (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    code character varying(50) NOT NULL,
    discount_type character varying(20) NOT NULL,
    value numeric(10,2) NOT NULL,
    min_purchase numeric(10,2) DEFAULT 0,
    is_active boolean DEFAULT true,
    valid_from timestamp with time zone,
    valid_to timestamp with time zone,
    usage_limit bigint DEFAULT 100,
    used_count bigint DEFAULT 0,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: crm_settings; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.crm_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    points_per_currency numeric(5,2) DEFAULT 1,
    redemption_rate numeric(5,2) DEFAULT 0.01,
    version bigint DEFAULT 1 NOT NULL,
    monthly_sales_goal numeric(12,2) DEFAULT 50000
);


--
-- Name: customer_feedbacks; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.customer_feedbacks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    customer_id uuid NOT NULL,
    order_id uuid,
    rating bigint DEFAULT 5,
    comment text,
    is_public boolean DEFAULT false,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: customer_segments; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.customer_segments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    criteria_json jsonb
);


--
-- Name: customer_tiers; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.customer_tiers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(50) NOT NULL,
    min_spend numeric(12,2) DEFAULT 0,
    discount_percentage numeric(5,2) DEFAULT 0,
    color character varying(20) DEFAULT 'blue'::character varying,
    icon character varying(50) DEFAULT 'star'::character varying,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: customers; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.customers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(200) NOT NULL,
    phone text,
    email text,
    address text,
    tier_id uuid,
    total_spend numeric(12,2) DEFAULT 0,
    order_count bigint DEFAULT 0,
    loyalty_pts numeric(10,2) DEFAULT 0,
    store_credit numeric(12,2) DEFAULT 0,
    debt_balance numeric(12,2) DEFAULT 0,
    accepts_marketing boolean DEFAULT true,
    last_visit timestamp with time zone,
    date_of_birth timestamp with time zone,
    notes text,
    customer_type character varying(20) DEFAULT 'retail'::character varying,
    version bigint DEFAULT 1 NOT NULL,
    password_hash character varying(255),
    is_registered boolean DEFAULT false
);


--
-- Name: delivery_drivers; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.delivery_drivers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    phone character varying(20) NOT NULL,
    vehicle_info character varying(100),
    current_status character varying(20) DEFAULT 'available'::character varying,
    lat numeric(10,8),
    lng numeric(11,8)
);


--
-- Name: delivery_orders; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.delivery_orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    order_id uuid NOT NULL,
    driver_id uuid,
    status character varying(20) DEFAULT 'pending'::character varying,
    tracking_link character varying(255),
    delivery_notes text,
    delivery_fee numeric(12,2) DEFAULT 0
);


--
-- Name: dining_tables; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.dining_tables (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    name character varying(50) NOT NULL,
    capacity bigint DEFAULT 4,
    status character varying(20) DEFAULT 'available'::character varying,
    qr_code_url character varying(512),
    position_x bigint DEFAULT 0,
    position_y bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: discount_codes; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.discount_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    type text NOT NULL,
    value numeric NOT NULL,
    status text DEFAULT 'active'::text,
    max_uses bigint,
    current_uses bigint DEFAULT 0,
    valid_from timestamp with time zone,
    valid_until timestamp with time zone,
    min_order_value numeric,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    points_required bigint
);


--
-- Name: domains; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.domains (
    id bigint NOT NULL,
    tenant_id uuid NOT NULL,
    domain character varying(253) NOT NULL,
    is_primary boolean DEFAULT false,
    is_verified boolean DEFAULT false,
    verification_token uuid DEFAULT gen_random_uuid(),
    dns_checked_at timestamp with time zone,
    created_at timestamp with time zone
);


--
-- Name: domains_id_seq; Type: SEQUENCE; Schema: thinkce; Owner: -
--

CREATE SEQUENCE thinkce.domains_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: domains_id_seq; Type: SEQUENCE OWNED BY; Schema: thinkce; Owner: -
--

ALTER SEQUENCE thinkce.domains_id_seq OWNED BY thinkce.domains.id;


--
-- Name: expense_categories; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.expense_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(100) NOT NULL,
    type character varying(20) DEFAULT 'variable'::character varying,
    description text,
    version bigint DEFAULT 1 NOT NULL,
    monthly_budget numeric(12,2) DEFAULT 0
);


--
-- Name: expenses; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.expenses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    category_id uuid NOT NULL,
    amount numeric(12,2) NOT NULL,
    date timestamp with time zone,
    description text,
    receipt_url character varying(512),
    created_by_id uuid,
    version bigint DEFAULT 1 NOT NULL,
    is_recurring boolean DEFAULT false,
    recurrence_interval character varying(20) DEFAULT ''::character varying
);


--
-- Name: external_systems; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.external_systems (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    developer_id uuid NOT NULL,
    name character varying(150) NOT NULL,
    description text,
    client_id uuid DEFAULT gen_random_uuid(),
    client_secret_hash character varying(128),
    redirect_uris jsonb,
    webhook_url character varying(512),
    icon character varying(50) DEFAULT 'rocket_launch'::character varying,
    is_public boolean DEFAULT false,
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: gift_cards; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.gift_cards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    code character varying(50) NOT NULL,
    initial_balance numeric(10,2) NOT NULL,
    current_balance numeric(10,2) NOT NULL,
    purchaser_id uuid,
    recipient_email character varying(254),
    expires_at timestamp with time zone,
    is_active boolean DEFAULT true,
    status character varying(20) DEFAULT 'active'::character varying,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: honeypot_attempts; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.honeypot_attempts (
    id bigint NOT NULL,
    username character varying(255),
    password character varying(255),
    ip_address character varying(45),
    user_agent text,
    path character varying(255) DEFAULT '/admin/'::character varying,
    "timestamp" timestamp with time zone
);


--
-- Name: honeypot_attempts_id_seq; Type: SEQUENCE; Schema: thinkce; Owner: -
--

CREATE SEQUENCE thinkce.honeypot_attempts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: honeypot_attempts_id_seq; Type: SEQUENCE OWNED BY; Schema: thinkce; Owner: -
--

ALTER SEQUENCE thinkce.honeypot_attempts_id_seq OWNED BY thinkce.honeypot_attempts.id;


--
-- Name: journal_entries; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.journal_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    reference_id uuid,
    reference_type character varying(50),
    description text
);


--
-- Name: kds_tickets; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.kds_tickets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    order_id uuid NOT NULL,
    table_id uuid,
    status character varying(20) DEFAULT 'pending'::character varying,
    kitchen_notes text,
    is_rush boolean DEFAULT false,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: leave_requests; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.leave_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    staff_id uuid NOT NULL,
    leave_type character varying(20) NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    reason text,
    status character varying(20) DEFAULT 'pending'::character varying,
    reviewed_by_id uuid,
    reviewed_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: ledger_accounts; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.ledger_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    name character varying(100) NOT NULL,
    type character varying(50) NOT NULL,
    code character varying(20),
    description text
);


--
-- Name: ledger_lines; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.ledger_lines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    journal_entry_id uuid NOT NULL,
    account_id uuid NOT NULL,
    amount numeric(12,2) NOT NULL,
    is_debit boolean NOT NULL
);


--
-- Name: loyalty_transactions; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.loyalty_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    order_id uuid,
    points numeric(10,2) NOT NULL,
    transaction_type character varying(20) NOT NULL,
    description text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: marketing_campaigns; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.marketing_campaigns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(100) NOT NULL,
    campaign_type character varying(10) DEFAULT 'email'::character varying,
    subject character varying(255),
    message text NOT NULL,
    coupon_code character varying(50),
    status character varying(20) DEFAULT 'draft'::character varying,
    target_tier_id uuid,
    is_automated boolean DEFAULT false,
    trigger_event character varying(30) DEFAULT 'manual'::character varying,
    scheduled_at timestamp with time zone,
    sent_at timestamp with time zone,
    last_run_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    segment_id uuid,
    open_count bigint DEFAULT 0,
    click_count bigint DEFAULT 0,
    conversion_count bigint DEFAULT 0,
    revenue_generated numeric(12,2) DEFAULT 0
);


--
-- Name: newsletter_subscriptions; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.newsletter_subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email character varying(255) NOT NULL,
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: notification_settings; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.notification_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid NOT NULL,
    email_notifications boolean DEFAULT true,
    low_stock_alerts boolean DEFAULT true,
    sales_reports boolean DEFAULT true,
    security_alerts boolean DEFAULT true,
    system_alerts boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: notifications; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    message text NOT NULL,
    link character varying(255),
    is_read boolean DEFAULT false,
    notification_type character varying(20) DEFAULT 'info'::character varying,
    category character varying(20) DEFAULT 'general'::character varying,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: order_items; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    quantity numeric(10,3) NOT NULL,
    unit_price numeric(12,2) NOT NULL,
    discount numeric(12,2) DEFAULT 0,
    total numeric(12,2) NOT NULL,
    cost_price numeric(12,2) DEFAULT 0,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: orders; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    order_number character varying(50) NOT NULL,
    customer_id uuid,
    cashier_id uuid,
    subtotal numeric(12,2) DEFAULT 0,
    tax numeric(12,2) DEFAULT 0,
    discount numeric(12,2) DEFAULT 0,
    total numeric(12,2) NOT NULL,
    amount_paid numeric(12,2) DEFAULT 0,
    status character varying(20) DEFAULT 'completed'::character varying,
    payment_status character varying(20) DEFAULT 'paid'::character varying,
    payment_method character varying(20),
    order_type character varying(20) DEFAULT 'in_store'::character varying,
    notes text,
    receipt_token character varying(64),
    version bigint DEFAULT 1
);


--
-- Name: payment_methods; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.payment_methods (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(100) NOT NULL,
    provider character varying(20) NOT NULL,
    is_active boolean DEFAULT true,
    api_key_hint character varying(50),
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: payments; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_id uuid NOT NULL,
    payment_method_id uuid NOT NULL,
    amount numeric(12,2) NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    transaction_id character varying(255),
    metadata jsonb,
    error_message text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: payroll_periods; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.payroll_periods (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(100) NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    is_processed boolean DEFAULT false,
    processed_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: payroll_records; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.payroll_records (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    period_id uuid NOT NULL,
    staff_id uuid NOT NULL,
    base_salary_snapshot numeric(12,2) NOT NULL,
    total_commission numeric(12,2) DEFAULT 0,
    bonus numeric(12,2) DEFAULT 0,
    deductions numeric(12,2) DEFAULT 0,
    net_pay numeric(12,2) NOT NULL,
    is_paid boolean DEFAULT false,
    paid_at timestamp with time zone,
    payment_reference character varying(100),
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: print_jobs; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.print_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid NOT NULL,
    document_type character varying(50) NOT NULL,
    content text NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    printed_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: product_components; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.product_components (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    composite_product_id uuid NOT NULL,
    component_product_id uuid NOT NULL,
    quantity numeric(10,4) NOT NULL,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: product_histories; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.product_histories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id uuid NOT NULL,
    user_id uuid,
    field character varying(50),
    old_value text,
    new_value text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: product_image_galleries; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.product_image_galleries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id uuid NOT NULL,
    image_url character varying(512) NOT NULL,
    alt_text character varying(255),
    "order" bigint DEFAULT 0,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: product_reviews; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.product_reviews (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    rating bigint NOT NULL,
    comment text,
    is_visible boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: product_variants; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.product_variants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    sku character varying(100) NOT NULL,
    barcode character varying(100),
    price_override numeric(12,2),
    cost_override numeric(12,2),
    current_stock numeric(10,4) DEFAULT 0,
    attributes jsonb,
    image character varying(255),
    version bigint DEFAULT 1 NOT NULL,
    CONSTRAINT chk_product_variants_attributes CHECK ((jsonb_typeof(attributes) = 'object'::text))
);


--
-- Name: products; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    name character varying(200) NOT NULL,
    description text,
    sku character varying(100) NOT NULL,
    barcode character varying(100),
    category_id uuid,
    cost_price numeric(12,2) DEFAULT 0,
    selling_price numeric(12,2) NOT NULL,
    wholesale_price numeric(12,2) DEFAULT 0,
    track_inventory boolean DEFAULT true,
    current_stock numeric(10,4) DEFAULT 0,
    reorder_level numeric(10,4) DEFAULT 0,
    stock_unit character varying(50) DEFAULT 'pcs'::character varying,
    has_variants boolean DEFAULT false,
    is_composite boolean DEFAULT false,
    is_active boolean DEFAULT true,
    is_online boolean DEFAULT true,
    image character varying(255),
    supplier_id uuid,
    last_received_date timestamp with time zone,
    expiry_date timestamp with time zone,
    manufacturing_date timestamp with time zone,
    minimum_wholesale_quantity numeric(10,4) DEFAULT 1,
    batch_number character varying(100),
    invoice_waybill_number character varying(100),
    country_of_origin character varying(100),
    manufacturer_name character varying(200),
    manufacturer_address text,
    version bigint DEFAULT 1
);


--
-- Name: promotions; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.promotions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'draft'::text,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    description text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    points_required bigint
);


--
-- Name: purchase_order_items; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.purchase_order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    po_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    quantity_ordered numeric(10,4) NOT NULL,
    quantity_received numeric(10,4) DEFAULT 0,
    unit_cost numeric(12,2) NOT NULL,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: purchase_orders; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.purchase_orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    po_number character varying(50) NOT NULL,
    supplier_id uuid NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying,
    total_amount numeric(12,2) DEFAULT 0,
    expected_date timestamp with time zone,
    notes text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: quotation_items; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.quotation_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    quotation_id uuid NOT NULL,
    product_id uuid NOT NULL,
    quantity bigint DEFAULT 1,
    unit_price numeric(12,2) NOT NULL,
    discount numeric(12,2) DEFAULT 0,
    total_price numeric(12,2) NOT NULL,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: quotations; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.quotations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    customer_id uuid NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying,
    quote_number character varying(30) NOT NULL,
    subtotal numeric(12,2) DEFAULT 0,
    tax_amount numeric(12,2) DEFAULT 0,
    total_amount numeric(12,2) DEFAULT 0,
    notes text,
    internal_notes text,
    valid_until timestamp with time zone,
    created_by_id uuid,
    reviewed_by_id uuid,
    converted_order_id uuid,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: return_items; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.return_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    return_id uuid NOT NULL,
    product_id uuid,
    quantity numeric(10,3) DEFAULT 1,
    condition character varying(20) DEFAULT 'opened'::character varying,
    restock boolean DEFAULT false,
    unit_price numeric(12,2) NOT NULL,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: returns; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.returns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    order_id uuid NOT NULL,
    customer_id uuid,
    reason character varying(50) NOT NULL,
    reason_detail text,
    status character varying(20) DEFAULT 'pending'::character varying,
    refund_method character varying(20) DEFAULT 'original'::character varying,
    refund_amount numeric(12,2) DEFAULT 0,
    restocking_fee numeric(12,2) DEFAULT 0,
    created_by_id uuid,
    approved_by_id uuid,
    approved_at timestamp with time zone,
    completed_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: service_categories; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.service_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(100) NOT NULL,
    icon character varying(50) DEFAULT 'spa'::character varying,
    color character varying(20) DEFAULT 'purple'::character varying,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: service_commission_records; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.service_commission_records (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    staff_member_id uuid NOT NULL,
    rule_id uuid,
    order_id uuid NOT NULL,
    amount numeric(10,2) NOT NULL,
    is_paid boolean DEFAULT false,
    paid_at timestamp with time zone,
    notes text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: service_commission_rules; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.service_commission_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    staff_member_id uuid NOT NULL,
    commission_type character varying(20) DEFAULT 'percentage'::character varying,
    value numeric(8,2) NOT NULL,
    applies_to character varying(20) DEFAULT 'all_sales'::character varying,
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: services; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.services (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    category_id uuid,
    name character varying(200) NOT NULL,
    description text,
    duration_minutes bigint DEFAULT 30,
    price numeric(10,2) NOT NULL,
    default_staff_id uuid,
    image character varying(512),
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: shifts; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.shifts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid NOT NULL,
    user_id uuid NOT NULL,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: split_bill_groups; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.split_bill_groups (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    table_id uuid,
    original_order_id uuid,
    notes text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: stock_batches; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.stock_batches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    branch_id uuid NOT NULL,
    product_id uuid NOT NULL,
    batch_number character varying(100) NOT NULL,
    quantity numeric(10,4) NOT NULL,
    expiry_date timestamp with time zone,
    manufacture_date timestamp with time zone
);


--
-- Name: stock_movements; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.stock_movements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    branch_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    quantity numeric(10,4) NOT NULL,
    previous_stock numeric(10,4) NOT NULL,
    new_stock numeric(10,4) NOT NULL,
    reason character varying(50) NOT NULL,
    reference_id character varying(100),
    user_id uuid,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: stock_transfer_items; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.stock_transfer_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    transfer_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    quantity numeric(10,4) NOT NULL,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: stock_transfers; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.stock_transfers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    reference_no character varying(50) NOT NULL,
    from_branch_id uuid NOT NULL,
    to_branch_id uuid NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying,
    notes text,
    created_by_id uuid,
    shipped_at timestamp with time zone,
    received_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: stocktake_entries; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.stocktake_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    session_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    expected_stock numeric(10,4) NOT NULL,
    actual_stock numeric(10,4) NOT NULL,
    difference numeric(10,4) NOT NULL
);


--
-- Name: stocktake_sessions; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.stocktake_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    branch_id uuid,
    name character varying(100) NOT NULL,
    status character varying(20) DEFAULT 'in_progress'::character varying,
    notes text,
    created_by_id uuid,
    completed_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    access_token uuid
);


--
-- Name: storefront_settings; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.storefront_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    default_branch_id uuid,
    is_active boolean DEFAULT false,
    slug character varying(100),
    store_view_type character varying(20) DEFAULT 'branch'::character varying,
    store_name character varying(100),
    banner_image character varying(255),
    logo_image character varying(255),
    primary_color character varying(7) DEFAULT '#3b82f6'::character varying,
    welcome_message text,
    about_text text,
    allow_pickup boolean DEFAULT true,
    allow_delivery boolean DEFAULT false,
    delivery_fee numeric(8,2) DEFAULT 0,
    min_order_amount numeric(10,2) DEFAULT 0,
    enable_stripe boolean DEFAULT false,
    enable_paystack boolean DEFAULT false,
    enable_mobile_money boolean DEFAULT false,
    version bigint DEFAULT 1 NOT NULL,
    paystack_public_key character varying(255),
    paystack_subaccount_code character varying(100),
    flash_sale_end_time timestamp with time zone
);


--
-- Name: supplier_ledger_entries; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.supplier_ledger_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    supplier_id uuid NOT NULL,
    entry_type character varying(50) NOT NULL,
    reference_id character varying(100),
    amount numeric(12,2) NOT NULL,
    balance numeric(12,2) NOT NULL,
    notes text
);


--
-- Name: supplier_products; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.supplier_products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL,
    supplier_id uuid NOT NULL,
    product_id uuid NOT NULL,
    supplier_sku character varying(100),
    unit_cost numeric(12,2) DEFAULT 0 NOT NULL
);


--
-- Name: supplier_profiles; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.supplier_profiles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id uuid NOT NULL,
    supplier_id uuid NOT NULL,
    role character varying(50) DEFAULT 'contact'::character varying,
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: suppliers; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.suppliers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(200) NOT NULL,
    contact_person character varying(150),
    email text,
    phone text,
    address text,
    tax_number character varying(50),
    payment_terms text,
    notes text,
    credit_balance numeric(12,2) DEFAULT 0,
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: tax_configurations; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.tax_configurations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tax_type character varying(20) DEFAULT 'sales_tax'::character varying,
    tax_rate numeric(5,2) DEFAULT 0,
    tax_number character varying(50),
    include_tax_in_prices boolean DEFAULT false,
    is_active boolean DEFAULT true,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: tenant_integrations; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.tenant_integrations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    provider character varying(50) NOT NULL,
    access_token text,
    refresh_token text,
    expires_at timestamp with time zone,
    config jsonb,
    is_active boolean DEFAULT true,
    last_sync_at timestamp with time zone,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: webhook_endpoints; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.webhook_endpoints (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    external_system_id uuid,
    url character varying(512) NOT NULL,
    secret character varying(64) NOT NULL,
    is_active boolean DEFAULT true,
    events jsonb,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: webhook_events; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.webhook_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    endpoint_id uuid NOT NULL,
    event_type character varying(50) NOT NULL,
    payload jsonb,
    signature character varying(128),
    status_code bigint,
    response_body text,
    error_message text,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: wishlists; Type: TABLE; Schema: thinkce; Owner: -
--

CREATE TABLE thinkce.wishlists (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    customer_id uuid NOT NULL,
    product_id uuid NOT NULL,
    version bigint DEFAULT 1 NOT NULL
);


--
-- Name: billing_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_settings ALTER COLUMN id SET DEFAULT nextval('public.billing_settings_id_seq'::regclass);


--
-- Name: domains id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.domains ALTER COLUMN id SET DEFAULT nextval('public.domains_id_seq'::regclass);


--
-- Name: domains id; Type: DEFAULT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.domains ALTER COLUMN id SET DEFAULT nextval('softivite.domains_id_seq'::regclass);


--
-- Name: honeypot_attempts id; Type: DEFAULT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.honeypot_attempts ALTER COLUMN id SET DEFAULT nextval('softivite.honeypot_attempts_id_seq'::regclass);


--
-- Name: domains id; Type: DEFAULT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.domains ALTER COLUMN id SET DEFAULT nextval('thinkce.domains_id_seq'::regclass);


--
-- Name: honeypot_attempts id; Type: DEFAULT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.honeypot_attempts ALTER COLUMN id SET DEFAULT nextval('thinkce.honeypot_attempts_id_seq'::regclass);


--
-- Name: admin_roles admin_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_roles
    ADD CONSTRAINT admin_roles_pkey PRIMARY KEY (id);


--
-- Name: admin_users admin_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT admin_users_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: billing_payments billing_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_payments
    ADD CONSTRAINT billing_payments_pkey PRIMARY KEY (id);


--
-- Name: billing_settings billing_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_settings
    ADD CONSTRAINT billing_settings_pkey PRIMARY KEY (id);


--
-- Name: blog_posts blog_posts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blog_posts
    ADD CONSTRAINT blog_posts_pkey PRIMARY KEY (id);


--
-- Name: branches branches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.branches
    ADD CONSTRAINT branches_pkey PRIMARY KEY (id);


--
-- Name: broadcasts broadcasts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.broadcasts
    ADD CONSTRAINT broadcasts_pkey PRIMARY KEY (id);


--
-- Name: cross_tenant_audit_logs cross_tenant_audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cross_tenant_audit_logs
    ADD CONSTRAINT cross_tenant_audit_logs_pkey PRIMARY KEY (id);


--
-- Name: domains domains_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.domains
    ADD CONSTRAINT domains_pkey PRIMARY KEY (id);


--
-- Name: ip_allowlists ip_allowlists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ip_allowlists
    ADD CONSTRAINT ip_allowlists_pkey PRIMARY KEY (id);


--
-- Name: legal_documents legal_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.legal_documents
    ADD CONSTRAINT legal_documents_pkey PRIMARY KEY (id);


--
-- Name: master_api_keys master_api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.master_api_keys
    ADD CONSTRAINT master_api_keys_pkey PRIMARY KEY (id);


--
-- Name: plan_features plan_features_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features
    ADD CONSTRAINT plan_features_pkey PRIMARY KEY (id);


--
-- Name: plans plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_pkey PRIMARY KEY (id);


--
-- Name: pricing_plans pricing_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pricing_plans
    ADD CONSTRAINT pricing_plans_pkey PRIMARY KEY (id);


--
-- Name: promo_codes promo_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_pkey PRIMARY KEY (id);


--
-- Name: referral_rewards referral_rewards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_rewards
    ADD CONSTRAINT referral_rewards_pkey PRIMARY KEY (id);


--
-- Name: seo_settings seo_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seo_settings
    ADD CONSTRAINT seo_settings_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: tenant_metrics tenant_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_metrics
    ADD CONSTRAINT tenant_metrics_pkey PRIMARY KEY (id);


--
-- Name: tenants tenants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: abandoned_carts abandoned_carts_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.abandoned_carts
    ADD CONSTRAINT abandoned_carts_pkey PRIMARY KEY (id);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: api_request_logs api_request_logs_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.api_request_logs
    ADD CONSTRAINT api_request_logs_pkey PRIMARY KEY (id);


--
-- Name: appointments appointments_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.appointments
    ADD CONSTRAINT appointments_pkey PRIMARY KEY (id);


--
-- Name: attendances attendances_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.attendances
    ADD CONSTRAINT attendances_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: branches branches_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.branches
    ADD CONSTRAINT branches_pkey PRIMARY KEY (id);


--
-- Name: cash_drawer_sessions cash_drawer_sessions_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.cash_drawer_sessions
    ADD CONSTRAINT cash_drawer_sessions_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: commission_rules commission_rules_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.commission_rules
    ADD CONSTRAINT commission_rules_pkey PRIMARY KEY (id);


--
-- Name: coupons coupons_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.coupons
    ADD CONSTRAINT coupons_pkey PRIMARY KEY (id);


--
-- Name: crm_settings crm_settings_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.crm_settings
    ADD CONSTRAINT crm_settings_pkey PRIMARY KEY (id);


--
-- Name: customer_feedbacks customer_feedbacks_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.customer_feedbacks
    ADD CONSTRAINT customer_feedbacks_pkey PRIMARY KEY (id);


--
-- Name: customer_segments customer_segments_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.customer_segments
    ADD CONSTRAINT customer_segments_pkey PRIMARY KEY (id);


--
-- Name: customer_tiers customer_tiers_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.customer_tiers
    ADD CONSTRAINT customer_tiers_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: delivery_drivers delivery_drivers_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.delivery_drivers
    ADD CONSTRAINT delivery_drivers_pkey PRIMARY KEY (id);


--
-- Name: delivery_orders delivery_orders_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.delivery_orders
    ADD CONSTRAINT delivery_orders_pkey PRIMARY KEY (id);


--
-- Name: dining_tables dining_tables_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.dining_tables
    ADD CONSTRAINT dining_tables_pkey PRIMARY KEY (id);


--
-- Name: discount_codes discount_codes_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.discount_codes
    ADD CONSTRAINT discount_codes_pkey PRIMARY KEY (id);


--
-- Name: domains domains_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.domains
    ADD CONSTRAINT domains_pkey PRIMARY KEY (id);


--
-- Name: expense_categories expense_categories_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.expense_categories
    ADD CONSTRAINT expense_categories_pkey PRIMARY KEY (id);


--
-- Name: expenses expenses_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.expenses
    ADD CONSTRAINT expenses_pkey PRIMARY KEY (id);


--
-- Name: external_systems external_systems_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.external_systems
    ADD CONSTRAINT external_systems_pkey PRIMARY KEY (id);


--
-- Name: gift_cards gift_cards_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.gift_cards
    ADD CONSTRAINT gift_cards_pkey PRIMARY KEY (id);


--
-- Name: honeypot_attempts honeypot_attempts_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.honeypot_attempts
    ADD CONSTRAINT honeypot_attempts_pkey PRIMARY KEY (id);


--
-- Name: journal_entries journal_entries_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.journal_entries
    ADD CONSTRAINT journal_entries_pkey PRIMARY KEY (id);


--
-- Name: kds_tickets kds_tickets_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.kds_tickets
    ADD CONSTRAINT kds_tickets_pkey PRIMARY KEY (id);


--
-- Name: leave_requests leave_requests_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.leave_requests
    ADD CONSTRAINT leave_requests_pkey PRIMARY KEY (id);


--
-- Name: ledger_accounts ledger_accounts_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.ledger_accounts
    ADD CONSTRAINT ledger_accounts_pkey PRIMARY KEY (id);


--
-- Name: ledger_lines ledger_lines_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.ledger_lines
    ADD CONSTRAINT ledger_lines_pkey PRIMARY KEY (id);


--
-- Name: loyalty_transactions loyalty_transactions_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.loyalty_transactions
    ADD CONSTRAINT loyalty_transactions_pkey PRIMARY KEY (id);


--
-- Name: marketing_campaigns marketing_campaigns_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.marketing_campaigns
    ADD CONSTRAINT marketing_campaigns_pkey PRIMARY KEY (id);


--
-- Name: newsletter_subscriptions newsletter_subscriptions_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.newsletter_subscriptions
    ADD CONSTRAINT newsletter_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: notification_settings notification_settings_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.notification_settings
    ADD CONSTRAINT notification_settings_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: payment_methods payment_methods_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.payment_methods
    ADD CONSTRAINT payment_methods_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: payroll_periods payroll_periods_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.payroll_periods
    ADD CONSTRAINT payroll_periods_pkey PRIMARY KEY (id);


--
-- Name: payroll_records payroll_records_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.payroll_records
    ADD CONSTRAINT payroll_records_pkey PRIMARY KEY (id);


--
-- Name: print_jobs print_jobs_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.print_jobs
    ADD CONSTRAINT print_jobs_pkey PRIMARY KEY (id);


--
-- Name: product_components product_components_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.product_components
    ADD CONSTRAINT product_components_pkey PRIMARY KEY (id);


--
-- Name: product_histories product_histories_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.product_histories
    ADD CONSTRAINT product_histories_pkey PRIMARY KEY (id);


--
-- Name: product_image_galleries product_image_galleries_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.product_image_galleries
    ADD CONSTRAINT product_image_galleries_pkey PRIMARY KEY (id);


--
-- Name: product_reviews product_reviews_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.product_reviews
    ADD CONSTRAINT product_reviews_pkey PRIMARY KEY (id);


--
-- Name: product_variants product_variants_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.product_variants
    ADD CONSTRAINT product_variants_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: promotions promotions_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.promotions
    ADD CONSTRAINT promotions_pkey PRIMARY KEY (id);


--
-- Name: purchase_order_items purchase_order_items_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.purchase_order_items
    ADD CONSTRAINT purchase_order_items_pkey PRIMARY KEY (id);


--
-- Name: purchase_orders purchase_orders_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.purchase_orders
    ADD CONSTRAINT purchase_orders_pkey PRIMARY KEY (id);


--
-- Name: quotation_items quotation_items_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.quotation_items
    ADD CONSTRAINT quotation_items_pkey PRIMARY KEY (id);


--
-- Name: quotations quotations_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.quotations
    ADD CONSTRAINT quotations_pkey PRIMARY KEY (id);


--
-- Name: return_items return_items_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.return_items
    ADD CONSTRAINT return_items_pkey PRIMARY KEY (id);


--
-- Name: returns returns_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.returns
    ADD CONSTRAINT returns_pkey PRIMARY KEY (id);


--
-- Name: service_categories service_categories_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.service_categories
    ADD CONSTRAINT service_categories_pkey PRIMARY KEY (id);


--
-- Name: service_commission_records service_commission_records_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.service_commission_records
    ADD CONSTRAINT service_commission_records_pkey PRIMARY KEY (id);


--
-- Name: service_commission_rules service_commission_rules_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.service_commission_rules
    ADD CONSTRAINT service_commission_rules_pkey PRIMARY KEY (id);


--
-- Name: services services_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.services
    ADD CONSTRAINT services_pkey PRIMARY KEY (id);


--
-- Name: shift_swap_requests shift_swap_requests_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.shift_swap_requests
    ADD CONSTRAINT shift_swap_requests_pkey PRIMARY KEY (id);


--
-- Name: shifts shifts_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.shifts
    ADD CONSTRAINT shifts_pkey PRIMARY KEY (id);


--
-- Name: split_bill_groups split_bill_groups_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.split_bill_groups
    ADD CONSTRAINT split_bill_groups_pkey PRIMARY KEY (id);


--
-- Name: staff_achievements staff_achievements_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.staff_achievements
    ADD CONSTRAINT staff_achievements_pkey PRIMARY KEY (id);


--
-- Name: stock_batches stock_batches_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_batches
    ADD CONSTRAINT stock_batches_pkey PRIMARY KEY (id);


--
-- Name: stock_movements stock_movements_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_movements
    ADD CONSTRAINT stock_movements_pkey PRIMARY KEY (id);


--
-- Name: stock_transfer_items stock_transfer_items_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_transfer_items
    ADD CONSTRAINT stock_transfer_items_pkey PRIMARY KEY (id);


--
-- Name: stock_transfers stock_transfers_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_transfers
    ADD CONSTRAINT stock_transfers_pkey PRIMARY KEY (id);


--
-- Name: stocktake_entries stocktake_entries_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stocktake_entries
    ADD CONSTRAINT stocktake_entries_pkey PRIMARY KEY (id);


--
-- Name: stocktake_sessions stocktake_sessions_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stocktake_sessions
    ADD CONSTRAINT stocktake_sessions_pkey PRIMARY KEY (id);


--
-- Name: storefront_settings storefront_settings_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.storefront_settings
    ADD CONSTRAINT storefront_settings_pkey PRIMARY KEY (id);


--
-- Name: supplier_ledger_entries supplier_ledger_entries_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.supplier_ledger_entries
    ADD CONSTRAINT supplier_ledger_entries_pkey PRIMARY KEY (id);


--
-- Name: supplier_products supplier_products_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.supplier_products
    ADD CONSTRAINT supplier_products_pkey PRIMARY KEY (id);


--
-- Name: supplier_profiles supplier_profiles_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.supplier_profiles
    ADD CONSTRAINT supplier_profiles_pkey PRIMARY KEY (id);


--
-- Name: suppliers suppliers_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.suppliers
    ADD CONSTRAINT suppliers_pkey PRIMARY KEY (id);


--
-- Name: tax_configurations tax_configurations_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.tax_configurations
    ADD CONSTRAINT tax_configurations_pkey PRIMARY KEY (id);


--
-- Name: tenant_integrations tenant_integrations_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.tenant_integrations
    ADD CONSTRAINT tenant_integrations_pkey PRIMARY KEY (id);


--
-- Name: discount_codes uni_discount_codes_code; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.discount_codes
    ADD CONSTRAINT uni_discount_codes_code UNIQUE (code);


--
-- Name: webhook_endpoints webhook_endpoints_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_pkey PRIMARY KEY (id);


--
-- Name: webhook_events webhook_events_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.webhook_events
    ADD CONSTRAINT webhook_events_pkey PRIMARY KEY (id);


--
-- Name: wishlists wishlists_pkey; Type: CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.wishlists
    ADD CONSTRAINT wishlists_pkey PRIMARY KEY (id);


--
-- Name: abandoned_carts abandoned_carts_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.abandoned_carts
    ADD CONSTRAINT abandoned_carts_pkey PRIMARY KEY (id);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: api_request_logs api_request_logs_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.api_request_logs
    ADD CONSTRAINT api_request_logs_pkey PRIMARY KEY (id);


--
-- Name: appointments appointments_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.appointments
    ADD CONSTRAINT appointments_pkey PRIMARY KEY (id);


--
-- Name: attendances attendances_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.attendances
    ADD CONSTRAINT attendances_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: branches branches_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.branches
    ADD CONSTRAINT branches_pkey PRIMARY KEY (id);


--
-- Name: cash_drawer_sessions cash_drawer_sessions_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.cash_drawer_sessions
    ADD CONSTRAINT cash_drawer_sessions_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: coupons coupons_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.coupons
    ADD CONSTRAINT coupons_pkey PRIMARY KEY (id);


--
-- Name: crm_settings crm_settings_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.crm_settings
    ADD CONSTRAINT crm_settings_pkey PRIMARY KEY (id);


--
-- Name: customer_feedbacks customer_feedbacks_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.customer_feedbacks
    ADD CONSTRAINT customer_feedbacks_pkey PRIMARY KEY (id);


--
-- Name: customer_segments customer_segments_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.customer_segments
    ADD CONSTRAINT customer_segments_pkey PRIMARY KEY (id);


--
-- Name: customer_tiers customer_tiers_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.customer_tiers
    ADD CONSTRAINT customer_tiers_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: delivery_drivers delivery_drivers_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.delivery_drivers
    ADD CONSTRAINT delivery_drivers_pkey PRIMARY KEY (id);


--
-- Name: delivery_orders delivery_orders_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.delivery_orders
    ADD CONSTRAINT delivery_orders_pkey PRIMARY KEY (id);


--
-- Name: dining_tables dining_tables_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.dining_tables
    ADD CONSTRAINT dining_tables_pkey PRIMARY KEY (id);


--
-- Name: discount_codes discount_codes_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.discount_codes
    ADD CONSTRAINT discount_codes_pkey PRIMARY KEY (id);


--
-- Name: domains domains_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.domains
    ADD CONSTRAINT domains_pkey PRIMARY KEY (id);


--
-- Name: expense_categories expense_categories_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.expense_categories
    ADD CONSTRAINT expense_categories_pkey PRIMARY KEY (id);


--
-- Name: expenses expenses_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.expenses
    ADD CONSTRAINT expenses_pkey PRIMARY KEY (id);


--
-- Name: external_systems external_systems_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.external_systems
    ADD CONSTRAINT external_systems_pkey PRIMARY KEY (id);


--
-- Name: gift_cards gift_cards_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.gift_cards
    ADD CONSTRAINT gift_cards_pkey PRIMARY KEY (id);


--
-- Name: honeypot_attempts honeypot_attempts_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.honeypot_attempts
    ADD CONSTRAINT honeypot_attempts_pkey PRIMARY KEY (id);


--
-- Name: journal_entries journal_entries_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.journal_entries
    ADD CONSTRAINT journal_entries_pkey PRIMARY KEY (id);


--
-- Name: kds_tickets kds_tickets_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.kds_tickets
    ADD CONSTRAINT kds_tickets_pkey PRIMARY KEY (id);


--
-- Name: leave_requests leave_requests_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.leave_requests
    ADD CONSTRAINT leave_requests_pkey PRIMARY KEY (id);


--
-- Name: ledger_accounts ledger_accounts_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.ledger_accounts
    ADD CONSTRAINT ledger_accounts_pkey PRIMARY KEY (id);


--
-- Name: ledger_lines ledger_lines_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.ledger_lines
    ADD CONSTRAINT ledger_lines_pkey PRIMARY KEY (id);


--
-- Name: loyalty_transactions loyalty_transactions_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.loyalty_transactions
    ADD CONSTRAINT loyalty_transactions_pkey PRIMARY KEY (id);


--
-- Name: marketing_campaigns marketing_campaigns_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.marketing_campaigns
    ADD CONSTRAINT marketing_campaigns_pkey PRIMARY KEY (id);


--
-- Name: newsletter_subscriptions newsletter_subscriptions_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.newsletter_subscriptions
    ADD CONSTRAINT newsletter_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: notification_settings notification_settings_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.notification_settings
    ADD CONSTRAINT notification_settings_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: payment_methods payment_methods_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.payment_methods
    ADD CONSTRAINT payment_methods_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: payroll_periods payroll_periods_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.payroll_periods
    ADD CONSTRAINT payroll_periods_pkey PRIMARY KEY (id);


--
-- Name: payroll_records payroll_records_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.payroll_records
    ADD CONSTRAINT payroll_records_pkey PRIMARY KEY (id);


--
-- Name: print_jobs print_jobs_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.print_jobs
    ADD CONSTRAINT print_jobs_pkey PRIMARY KEY (id);


--
-- Name: product_components product_components_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.product_components
    ADD CONSTRAINT product_components_pkey PRIMARY KEY (id);


--
-- Name: product_histories product_histories_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.product_histories
    ADD CONSTRAINT product_histories_pkey PRIMARY KEY (id);


--
-- Name: product_image_galleries product_image_galleries_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.product_image_galleries
    ADD CONSTRAINT product_image_galleries_pkey PRIMARY KEY (id);


--
-- Name: product_reviews product_reviews_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.product_reviews
    ADD CONSTRAINT product_reviews_pkey PRIMARY KEY (id);


--
-- Name: product_variants product_variants_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.product_variants
    ADD CONSTRAINT product_variants_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: promotions promotions_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.promotions
    ADD CONSTRAINT promotions_pkey PRIMARY KEY (id);


--
-- Name: purchase_order_items purchase_order_items_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.purchase_order_items
    ADD CONSTRAINT purchase_order_items_pkey PRIMARY KEY (id);


--
-- Name: purchase_orders purchase_orders_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.purchase_orders
    ADD CONSTRAINT purchase_orders_pkey PRIMARY KEY (id);


--
-- Name: quotation_items quotation_items_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.quotation_items
    ADD CONSTRAINT quotation_items_pkey PRIMARY KEY (id);


--
-- Name: quotations quotations_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.quotations
    ADD CONSTRAINT quotations_pkey PRIMARY KEY (id);


--
-- Name: return_items return_items_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.return_items
    ADD CONSTRAINT return_items_pkey PRIMARY KEY (id);


--
-- Name: returns returns_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.returns
    ADD CONSTRAINT returns_pkey PRIMARY KEY (id);


--
-- Name: service_categories service_categories_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.service_categories
    ADD CONSTRAINT service_categories_pkey PRIMARY KEY (id);


--
-- Name: service_commission_records service_commission_records_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.service_commission_records
    ADD CONSTRAINT service_commission_records_pkey PRIMARY KEY (id);


--
-- Name: service_commission_rules service_commission_rules_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.service_commission_rules
    ADD CONSTRAINT service_commission_rules_pkey PRIMARY KEY (id);


--
-- Name: services services_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.services
    ADD CONSTRAINT services_pkey PRIMARY KEY (id);


--
-- Name: shifts shifts_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.shifts
    ADD CONSTRAINT shifts_pkey PRIMARY KEY (id);


--
-- Name: split_bill_groups split_bill_groups_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.split_bill_groups
    ADD CONSTRAINT split_bill_groups_pkey PRIMARY KEY (id);


--
-- Name: stock_batches stock_batches_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_batches
    ADD CONSTRAINT stock_batches_pkey PRIMARY KEY (id);


--
-- Name: stock_movements stock_movements_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_movements
    ADD CONSTRAINT stock_movements_pkey PRIMARY KEY (id);


--
-- Name: stock_transfer_items stock_transfer_items_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_transfer_items
    ADD CONSTRAINT stock_transfer_items_pkey PRIMARY KEY (id);


--
-- Name: stock_transfers stock_transfers_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_transfers
    ADD CONSTRAINT stock_transfers_pkey PRIMARY KEY (id);


--
-- Name: stocktake_entries stocktake_entries_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stocktake_entries
    ADD CONSTRAINT stocktake_entries_pkey PRIMARY KEY (id);


--
-- Name: stocktake_sessions stocktake_sessions_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stocktake_sessions
    ADD CONSTRAINT stocktake_sessions_pkey PRIMARY KEY (id);


--
-- Name: storefront_settings storefront_settings_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.storefront_settings
    ADD CONSTRAINT storefront_settings_pkey PRIMARY KEY (id);


--
-- Name: supplier_ledger_entries supplier_ledger_entries_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.supplier_ledger_entries
    ADD CONSTRAINT supplier_ledger_entries_pkey PRIMARY KEY (id);


--
-- Name: supplier_products supplier_products_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.supplier_products
    ADD CONSTRAINT supplier_products_pkey PRIMARY KEY (id);


--
-- Name: supplier_profiles supplier_profiles_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.supplier_profiles
    ADD CONSTRAINT supplier_profiles_pkey PRIMARY KEY (id);


--
-- Name: suppliers suppliers_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.suppliers
    ADD CONSTRAINT suppliers_pkey PRIMARY KEY (id);


--
-- Name: tax_configurations tax_configurations_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.tax_configurations
    ADD CONSTRAINT tax_configurations_pkey PRIMARY KEY (id);


--
-- Name: tenant_integrations tenant_integrations_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.tenant_integrations
    ADD CONSTRAINT tenant_integrations_pkey PRIMARY KEY (id);


--
-- Name: discount_codes uni_discount_codes_code; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.discount_codes
    ADD CONSTRAINT uni_discount_codes_code UNIQUE (code);


--
-- Name: webhook_endpoints webhook_endpoints_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.webhook_endpoints
    ADD CONSTRAINT webhook_endpoints_pkey PRIMARY KEY (id);


--
-- Name: webhook_events webhook_events_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.webhook_events
    ADD CONSTRAINT webhook_events_pkey PRIMARY KEY (id);


--
-- Name: wishlists wishlists_pkey; Type: CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.wishlists
    ADD CONSTRAINT wishlists_pkey PRIMARY KEY (id);


--
-- Name: idx_admin_roles_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_roles_created_at ON public.admin_roles USING btree (created_at);


--
-- Name: idx_admin_roles_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_roles_deleted_at ON public.admin_roles USING btree (deleted_at);


--
-- Name: idx_admin_roles_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_admin_roles_name ON public.admin_roles USING btree (name);


--
-- Name: idx_admin_roles_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_roles_updated_at ON public.admin_roles USING btree (updated_at);


--
-- Name: idx_admin_roles_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_roles_version ON public.admin_roles USING btree (version);


--
-- Name: idx_admin_users_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_users_created_at ON public.admin_users USING btree (created_at);


--
-- Name: idx_admin_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_users_deleted_at ON public.admin_users USING btree (deleted_at);


--
-- Name: idx_admin_users_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_users_updated_at ON public.admin_users USING btree (updated_at);


--
-- Name: idx_admin_users_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_admin_users_user_id ON public.admin_users USING btree (user_id);


--
-- Name: idx_admin_users_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_users_version ON public.admin_users USING btree (version);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at);


--
-- Name: idx_audit_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_deleted_at ON public.audit_logs USING btree (deleted_at);


--
-- Name: idx_audit_logs_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_tenant_id ON public.audit_logs USING btree (tenant_id);


--
-- Name: idx_audit_logs_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_updated_at ON public.audit_logs USING btree (updated_at);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


--
-- Name: idx_audit_logs_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_version ON public.audit_logs USING btree (version);


--
-- Name: idx_billing_payments_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_billing_payments_created_at ON public.billing_payments USING btree (created_at);


--
-- Name: idx_billing_payments_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_billing_payments_deleted_at ON public.billing_payments USING btree (deleted_at);


--
-- Name: idx_billing_payments_subscription_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_billing_payments_subscription_id ON public.billing_payments USING btree (subscription_id);


--
-- Name: idx_billing_payments_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_billing_payments_updated_at ON public.billing_payments USING btree (updated_at);


--
-- Name: idx_billing_payments_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_billing_payments_version ON public.billing_payments USING btree (version);


--
-- Name: idx_blog_posts_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blog_posts_created_at ON public.blog_posts USING btree (created_at);


--
-- Name: idx_blog_posts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blog_posts_deleted_at ON public.blog_posts USING btree (deleted_at);


--
-- Name: idx_blog_posts_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_blog_posts_slug ON public.blog_posts USING btree (slug);


--
-- Name: idx_blog_posts_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blog_posts_updated_at ON public.blog_posts USING btree (updated_at);


--
-- Name: idx_blog_posts_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blog_posts_version ON public.blog_posts USING btree (version);


--
-- Name: idx_branches_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_branches_deleted_at ON public.branches USING btree (deleted_at);


--
-- Name: idx_branches_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_branches_tenant_id ON public.branches USING btree (tenant_id);


--
-- Name: idx_branches_unique_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_branches_unique_id ON public.branches USING btree (unique_id);


--
-- Name: idx_broadcasts_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_broadcasts_created_at ON public.broadcasts USING btree (created_at);


--
-- Name: idx_broadcasts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_broadcasts_deleted_at ON public.broadcasts USING btree (deleted_at);


--
-- Name: idx_broadcasts_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_broadcasts_updated_at ON public.broadcasts USING btree (updated_at);


--
-- Name: idx_broadcasts_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_broadcasts_version ON public.broadcasts USING btree (version);


--
-- Name: idx_cross_tenant_audit_logs_accessed_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cross_tenant_audit_logs_accessed_tenant_id ON public.cross_tenant_audit_logs USING btree (accessed_tenant_id);


--
-- Name: idx_cross_tenant_audit_logs_action_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cross_tenant_audit_logs_action_type ON public.cross_tenant_audit_logs USING btree (action_type);


--
-- Name: idx_cross_tenant_audit_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cross_tenant_audit_logs_created_at ON public.cross_tenant_audit_logs USING btree (created_at);


--
-- Name: idx_cross_tenant_audit_logs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cross_tenant_audit_logs_deleted_at ON public.cross_tenant_audit_logs USING btree (deleted_at);


--
-- Name: idx_cross_tenant_audit_logs_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cross_tenant_audit_logs_updated_at ON public.cross_tenant_audit_logs USING btree (updated_at);


--
-- Name: idx_cross_tenant_audit_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cross_tenant_audit_logs_user_id ON public.cross_tenant_audit_logs USING btree (user_id);


--
-- Name: idx_cross_tenant_audit_logs_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cross_tenant_audit_logs_version ON public.cross_tenant_audit_logs USING btree (version);


--
-- Name: idx_domains_domain; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_domains_domain ON public.domains USING btree (domain);


--
-- Name: idx_domains_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_domains_tenant_id ON public.domains USING btree (tenant_id);


--
-- Name: idx_ip_allowlists_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ip_allowlists_created_at ON public.ip_allowlists USING btree (created_at);


--
-- Name: idx_ip_allowlists_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ip_allowlists_deleted_at ON public.ip_allowlists USING btree (deleted_at);


--
-- Name: idx_ip_allowlists_ip_address; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_ip_allowlists_ip_address ON public.ip_allowlists USING btree (ip_address);


--
-- Name: idx_ip_allowlists_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ip_allowlists_updated_at ON public.ip_allowlists USING btree (updated_at);


--
-- Name: idx_ip_allowlists_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ip_allowlists_version ON public.ip_allowlists USING btree (version);


--
-- Name: idx_legal_documents_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_legal_documents_created_at ON public.legal_documents USING btree (created_at);


--
-- Name: idx_legal_documents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_legal_documents_deleted_at ON public.legal_documents USING btree (deleted_at);


--
-- Name: idx_legal_documents_type; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_legal_documents_type ON public.legal_documents USING btree (type);


--
-- Name: idx_legal_documents_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_legal_documents_updated_at ON public.legal_documents USING btree (updated_at);


--
-- Name: idx_legal_documents_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_legal_documents_version ON public.legal_documents USING btree (version);


--
-- Name: idx_master_api_keys_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_master_api_keys_created_at ON public.master_api_keys USING btree (created_at);


--
-- Name: idx_master_api_keys_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_master_api_keys_deleted_at ON public.master_api_keys USING btree (deleted_at);


--
-- Name: idx_master_api_keys_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_master_api_keys_key ON public.master_api_keys USING btree (key);


--
-- Name: idx_master_api_keys_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_master_api_keys_updated_at ON public.master_api_keys USING btree (updated_at);


--
-- Name: idx_master_api_keys_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_master_api_keys_version ON public.master_api_keys USING btree (version);


--
-- Name: idx_plan_features_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_created_at ON public.plan_features USING btree (created_at);


--
-- Name: idx_plan_features_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_deleted_at ON public.plan_features USING btree (deleted_at);


--
-- Name: idx_plan_features_plan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_plan_id ON public.plan_features USING btree (plan_id);


--
-- Name: idx_plan_features_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_updated_at ON public.plan_features USING btree (updated_at);


--
-- Name: idx_plan_features_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_version ON public.plan_features USING btree (version);


--
-- Name: idx_plans_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plans_created_at ON public.plans USING btree (created_at);


--
-- Name: idx_plans_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plans_deleted_at ON public.plans USING btree (deleted_at);


--
-- Name: idx_plans_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plans_updated_at ON public.plans USING btree (updated_at);


--
-- Name: idx_plans_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plans_version ON public.plans USING btree (version);


--
-- Name: idx_pricing_plans_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_plans_created_at ON public.pricing_plans USING btree (created_at);


--
-- Name: idx_pricing_plans_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_plans_deleted_at ON public.pricing_plans USING btree (deleted_at);


--
-- Name: idx_pricing_plans_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_pricing_plans_slug ON public.pricing_plans USING btree (slug);


--
-- Name: idx_pricing_plans_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_plans_updated_at ON public.pricing_plans USING btree (updated_at);


--
-- Name: idx_pricing_plans_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_plans_version ON public.pricing_plans USING btree (version);


--
-- Name: idx_promo_codes_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_promo_codes_code ON public.promo_codes USING btree (code);


--
-- Name: idx_promo_codes_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promo_codes_created_at ON public.promo_codes USING btree (created_at);


--
-- Name: idx_promo_codes_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promo_codes_deleted_at ON public.promo_codes USING btree (deleted_at);


--
-- Name: idx_promo_codes_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promo_codes_updated_at ON public.promo_codes USING btree (updated_at);


--
-- Name: idx_promo_codes_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promo_codes_version ON public.promo_codes USING btree (version);


--
-- Name: idx_public_tenants_referral_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_public_tenants_referral_code ON public.tenants USING btree (referral_code);


--
-- Name: idx_public_tenants_schema_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_public_tenants_schema_name ON public.tenants USING btree (schema_name);


--
-- Name: idx_public_tenants_subdomain; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_public_tenants_subdomain ON public.tenants USING btree (subdomain);


--
-- Name: idx_public_user_profiles_branch_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_public_user_profiles_branch_id ON public.user_profiles USING btree (branch_id);


--
-- Name: idx_public_user_profiles_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_public_user_profiles_deleted_at ON public.user_profiles USING btree (deleted_at);


--
-- Name: idx_public_user_profiles_role; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_public_user_profiles_role ON public.user_profiles USING btree (role);


--
-- Name: idx_public_user_profiles_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_public_user_profiles_tenant_id ON public.user_profiles USING btree (tenant_id);


--
-- Name: idx_public_user_profiles_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_public_user_profiles_user_id ON public.user_profiles USING btree (user_id);


--
-- Name: idx_public_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_public_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_public_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_public_users_email ON public.users USING btree (email);


--
-- Name: idx_public_users_reset_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_public_users_reset_token ON public.users USING btree (reset_token);


--
-- Name: idx_public_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_public_users_username ON public.users USING btree (username);


--
-- Name: idx_referral_rewards_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_referral_rewards_created_at ON public.referral_rewards USING btree (created_at);


--
-- Name: idx_referral_rewards_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_referral_rewards_deleted_at ON public.referral_rewards USING btree (deleted_at);


--
-- Name: idx_referral_rewards_referred_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_referral_rewards_referred_tenant_id ON public.referral_rewards USING btree (referred_tenant_id);


--
-- Name: idx_referral_rewards_referrer_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_referral_rewards_referrer_id ON public.referral_rewards USING btree (referrer_id);


--
-- Name: idx_referral_rewards_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_referral_rewards_updated_at ON public.referral_rewards USING btree (updated_at);


--
-- Name: idx_referral_rewards_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_referral_rewards_version ON public.referral_rewards USING btree (version);


--
-- Name: idx_seo_settings_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_seo_settings_created_at ON public.seo_settings USING btree (created_at);


--
-- Name: idx_seo_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_seo_settings_deleted_at ON public.seo_settings USING btree (deleted_at);


--
-- Name: idx_seo_settings_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_seo_settings_tenant_id ON public.seo_settings USING btree (tenant_id);


--
-- Name: idx_seo_settings_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_seo_settings_updated_at ON public.seo_settings USING btree (updated_at);


--
-- Name: idx_seo_settings_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_seo_settings_version ON public.seo_settings USING btree (version);


--
-- Name: idx_subscriptions_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscriptions_created_at ON public.subscriptions USING btree (created_at);


--
-- Name: idx_subscriptions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscriptions_deleted_at ON public.subscriptions USING btree (deleted_at);


--
-- Name: idx_subscriptions_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_subscriptions_tenant_id ON public.subscriptions USING btree (tenant_id);


--
-- Name: idx_subscriptions_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscriptions_updated_at ON public.subscriptions USING btree (updated_at);


--
-- Name: idx_subscriptions_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscriptions_version ON public.subscriptions USING btree (version);


--
-- Name: idx_tenant_metrics_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_metrics_created_at ON public.tenant_metrics USING btree (created_at);


--
-- Name: idx_tenant_metrics_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_metrics_deleted_at ON public.tenant_metrics USING btree (deleted_at);


--
-- Name: idx_tenant_metrics_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_metrics_tenant_id ON public.tenant_metrics USING btree (tenant_id);


--
-- Name: idx_tenant_metrics_updated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_metrics_updated_at ON public.tenant_metrics USING btree (updated_at);


--
-- Name: idx_tenant_metrics_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_metrics_version ON public.tenant_metrics USING btree (version);


--
-- Name: idx_abandoned_carts_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_abandoned_carts_created_at ON softivite.abandoned_carts USING btree (created_at);


--
-- Name: idx_abandoned_carts_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_abandoned_carts_deleted_at ON softivite.abandoned_carts USING btree (deleted_at);


--
-- Name: idx_abandoned_carts_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_abandoned_carts_updated_at ON softivite.abandoned_carts USING btree (updated_at);


--
-- Name: idx_abandoned_carts_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_abandoned_carts_version ON softivite.abandoned_carts USING btree (version);


--
-- Name: idx_api_keys_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_keys_branch_id ON softivite.api_keys USING btree (branch_id);


--
-- Name: idx_api_keys_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_keys_created_at ON softivite.api_keys USING btree (created_at);


--
-- Name: idx_api_keys_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_keys_deleted_at ON softivite.api_keys USING btree (deleted_at);


--
-- Name: idx_api_keys_key_prefix; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_keys_key_prefix ON softivite.api_keys USING btree (key_prefix);


--
-- Name: idx_api_keys_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_keys_updated_at ON softivite.api_keys USING btree (updated_at);


--
-- Name: idx_api_keys_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_keys_version ON softivite.api_keys USING btree (version);


--
-- Name: idx_api_request_logs_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_request_logs_created_at ON softivite.api_request_logs USING btree (created_at);


--
-- Name: idx_api_request_logs_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_request_logs_deleted_at ON softivite.api_request_logs USING btree (deleted_at);


--
-- Name: idx_api_request_logs_endpoint; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_request_logs_endpoint ON softivite.api_request_logs USING btree (endpoint);


--
-- Name: idx_api_request_logs_status_code; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_request_logs_status_code ON softivite.api_request_logs USING btree (status_code);


--
-- Name: idx_api_request_logs_tenant_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_request_logs_tenant_id ON softivite.api_request_logs USING btree (tenant_id);


--
-- Name: idx_api_request_logs_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_request_logs_updated_at ON softivite.api_request_logs USING btree (updated_at);


--
-- Name: idx_api_request_logs_user_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_request_logs_user_id ON softivite.api_request_logs USING btree (user_id);


--
-- Name: idx_api_request_logs_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_api_request_logs_version ON softivite.api_request_logs USING btree (version);


--
-- Name: idx_appointments_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_branch_id ON softivite.appointments USING btree (branch_id);


--
-- Name: idx_appointments_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_created_at ON softivite.appointments USING btree (created_at);


--
-- Name: idx_appointments_customer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_customer_id ON softivite.appointments USING btree (customer_id);


--
-- Name: idx_appointments_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_deleted_at ON softivite.appointments USING btree (deleted_at);


--
-- Name: idx_appointments_service_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_service_id ON softivite.appointments USING btree (service_id);


--
-- Name: idx_appointments_staff_member_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_staff_member_id ON softivite.appointments USING btree (staff_member_id);


--
-- Name: idx_appointments_start_time; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_start_time ON softivite.appointments USING btree (start_time);


--
-- Name: idx_appointments_status; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_status ON softivite.appointments USING btree (status);


--
-- Name: idx_appointments_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_updated_at ON softivite.appointments USING btree (updated_at);


--
-- Name: idx_appointments_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_appointments_version ON softivite.appointments USING btree (version);


--
-- Name: idx_attendances_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_attendances_branch_id ON softivite.attendances USING btree (branch_id);


--
-- Name: idx_attendances_clock_in; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_attendances_clock_in ON softivite.attendances USING btree (clock_in);


--
-- Name: idx_attendances_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_attendances_created_at ON softivite.attendances USING btree (created_at);


--
-- Name: idx_attendances_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_attendances_deleted_at ON softivite.attendances USING btree (deleted_at);


--
-- Name: idx_attendances_staff_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_attendances_staff_id ON softivite.attendances USING btree (staff_id);


--
-- Name: idx_attendances_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_attendances_updated_at ON softivite.attendances USING btree (updated_at);


--
-- Name: idx_attendances_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_attendances_version ON softivite.attendances USING btree (version);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_audit_logs_action ON softivite.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_audit_logs_created_at ON softivite.audit_logs USING btree (created_at);


--
-- Name: idx_audit_logs_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_audit_logs_deleted_at ON softivite.audit_logs USING btree (deleted_at);


--
-- Name: idx_audit_logs_tenant_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_audit_logs_tenant_id ON softivite.audit_logs USING btree (tenant_id);


--
-- Name: idx_audit_logs_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_audit_logs_updated_at ON softivite.audit_logs USING btree (updated_at);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON softivite.audit_logs USING btree (user_id);


--
-- Name: idx_audit_logs_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_audit_logs_version ON softivite.audit_logs USING btree (version);


--
-- Name: idx_branches_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_branches_deleted_at ON softivite.branches USING btree (deleted_at);


--
-- Name: idx_branches_tenant_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_branches_tenant_id ON softivite.branches USING btree (tenant_id);


--
-- Name: idx_branches_unique_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_branches_unique_id ON softivite.branches USING btree (unique_id);


--
-- Name: idx_cash_drawer_sessions_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_branch_id ON softivite.cash_drawer_sessions USING btree (branch_id);


--
-- Name: idx_cash_drawer_sessions_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_created_at ON softivite.cash_drawer_sessions USING btree (created_at);


--
-- Name: idx_cash_drawer_sessions_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_deleted_at ON softivite.cash_drawer_sessions USING btree (deleted_at);


--
-- Name: idx_cash_drawer_sessions_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_updated_at ON softivite.cash_drawer_sessions USING btree (updated_at);


--
-- Name: idx_cash_drawer_sessions_user_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_user_id ON softivite.cash_drawer_sessions USING btree (user_id);


--
-- Name: idx_cash_drawer_sessions_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_version ON softivite.cash_drawer_sessions USING btree (version);


--
-- Name: idx_categories_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_categories_created_at ON softivite.categories USING btree (created_at);


--
-- Name: idx_categories_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_categories_deleted_at ON softivite.categories USING btree (deleted_at);


--
-- Name: idx_categories_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_categories_updated_at ON softivite.categories USING btree (updated_at);


--
-- Name: idx_categories_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_categories_version ON softivite.categories USING btree (version);


--
-- Name: idx_commission_rules_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_commission_rules_branch_id ON softivite.commission_rules USING btree (branch_id);


--
-- Name: idx_commission_rules_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_commission_rules_created_at ON softivite.commission_rules USING btree (created_at);


--
-- Name: idx_commission_rules_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_commission_rules_deleted_at ON softivite.commission_rules USING btree (deleted_at);


--
-- Name: idx_commission_rules_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_commission_rules_updated_at ON softivite.commission_rules USING btree (updated_at);


--
-- Name: idx_commission_rules_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_commission_rules_version ON softivite.commission_rules USING btree (version);


--
-- Name: idx_coupons_code; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_coupons_code ON softivite.coupons USING btree (code);


--
-- Name: idx_coupons_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_coupons_created_at ON softivite.coupons USING btree (created_at);


--
-- Name: idx_coupons_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_coupons_deleted_at ON softivite.coupons USING btree (deleted_at);


--
-- Name: idx_coupons_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_coupons_updated_at ON softivite.coupons USING btree (updated_at);


--
-- Name: idx_coupons_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_coupons_version ON softivite.coupons USING btree (version);


--
-- Name: idx_crm_settings_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_crm_settings_created_at ON softivite.crm_settings USING btree (created_at);


--
-- Name: idx_crm_settings_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_crm_settings_deleted_at ON softivite.crm_settings USING btree (deleted_at);


--
-- Name: idx_crm_settings_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_crm_settings_updated_at ON softivite.crm_settings USING btree (updated_at);


--
-- Name: idx_crm_settings_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_crm_settings_version ON softivite.crm_settings USING btree (version);


--
-- Name: idx_customer_feedbacks_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_feedbacks_created_at ON softivite.customer_feedbacks USING btree (created_at);


--
-- Name: idx_customer_feedbacks_customer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_feedbacks_customer_id ON softivite.customer_feedbacks USING btree (customer_id);


--
-- Name: idx_customer_feedbacks_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_feedbacks_deleted_at ON softivite.customer_feedbacks USING btree (deleted_at);


--
-- Name: idx_customer_feedbacks_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_feedbacks_updated_at ON softivite.customer_feedbacks USING btree (updated_at);


--
-- Name: idx_customer_feedbacks_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_feedbacks_version ON softivite.customer_feedbacks USING btree (version);


--
-- Name: idx_customer_segments_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_segments_created_at ON softivite.customer_segments USING btree (created_at);


--
-- Name: idx_customer_segments_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_segments_deleted_at ON softivite.customer_segments USING btree (deleted_at);


--
-- Name: idx_customer_segments_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_segments_updated_at ON softivite.customer_segments USING btree (updated_at);


--
-- Name: idx_customer_segments_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_segments_version ON softivite.customer_segments USING btree (version);


--
-- Name: idx_customer_tiers_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_tiers_created_at ON softivite.customer_tiers USING btree (created_at);


--
-- Name: idx_customer_tiers_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_tiers_deleted_at ON softivite.customer_tiers USING btree (deleted_at);


--
-- Name: idx_customer_tiers_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_tiers_updated_at ON softivite.customer_tiers USING btree (updated_at);


--
-- Name: idx_customer_tiers_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customer_tiers_version ON softivite.customer_tiers USING btree (version);


--
-- Name: idx_customers_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customers_created_at ON softivite.customers USING btree (created_at);


--
-- Name: idx_customers_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customers_deleted_at ON softivite.customers USING btree (deleted_at);


--
-- Name: idx_customers_email; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customers_email ON softivite.customers USING btree (email);


--
-- Name: idx_customers_name; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customers_name ON softivite.customers USING btree (name);


--
-- Name: idx_customers_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customers_updated_at ON softivite.customers USING btree (updated_at);


--
-- Name: idx_customers_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_customers_version ON softivite.customers USING btree (version);


--
-- Name: idx_delivery_drivers_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_drivers_created_at ON softivite.delivery_drivers USING btree (created_at);


--
-- Name: idx_delivery_drivers_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_drivers_deleted_at ON softivite.delivery_drivers USING btree (deleted_at);


--
-- Name: idx_delivery_drivers_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_drivers_updated_at ON softivite.delivery_drivers USING btree (updated_at);


--
-- Name: idx_delivery_drivers_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_drivers_version ON softivite.delivery_drivers USING btree (version);


--
-- Name: idx_delivery_orders_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_orders_created_at ON softivite.delivery_orders USING btree (created_at);


--
-- Name: idx_delivery_orders_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_orders_deleted_at ON softivite.delivery_orders USING btree (deleted_at);


--
-- Name: idx_delivery_orders_driver_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_orders_driver_id ON softivite.delivery_orders USING btree (driver_id);


--
-- Name: idx_delivery_orders_order_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_delivery_orders_order_id ON softivite.delivery_orders USING btree (order_id);


--
-- Name: idx_delivery_orders_tracking_link; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_delivery_orders_tracking_link ON softivite.delivery_orders USING btree (tracking_link);


--
-- Name: idx_delivery_orders_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_orders_updated_at ON softivite.delivery_orders USING btree (updated_at);


--
-- Name: idx_delivery_orders_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_delivery_orders_version ON softivite.delivery_orders USING btree (version);


--
-- Name: idx_dining_tables_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_dining_tables_branch_id ON softivite.dining_tables USING btree (branch_id);


--
-- Name: idx_dining_tables_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_dining_tables_created_at ON softivite.dining_tables USING btree (created_at);


--
-- Name: idx_dining_tables_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_dining_tables_deleted_at ON softivite.dining_tables USING btree (deleted_at);


--
-- Name: idx_dining_tables_status; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_dining_tables_status ON softivite.dining_tables USING btree (status);


--
-- Name: idx_dining_tables_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_dining_tables_updated_at ON softivite.dining_tables USING btree (updated_at);


--
-- Name: idx_dining_tables_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_dining_tables_version ON softivite.dining_tables USING btree (version);


--
-- Name: idx_discount_codes_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_discount_codes_deleted_at ON softivite.discount_codes USING btree (deleted_at);


--
-- Name: idx_domains_domain; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_domains_domain ON softivite.domains USING btree (domain);


--
-- Name: idx_domains_tenant_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_domains_tenant_id ON softivite.domains USING btree (tenant_id);


--
-- Name: idx_expense_categories_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expense_categories_created_at ON softivite.expense_categories USING btree (created_at);


--
-- Name: idx_expense_categories_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expense_categories_deleted_at ON softivite.expense_categories USING btree (deleted_at);


--
-- Name: idx_expense_categories_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expense_categories_updated_at ON softivite.expense_categories USING btree (updated_at);


--
-- Name: idx_expense_categories_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expense_categories_version ON softivite.expense_categories USING btree (version);


--
-- Name: idx_expenses_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expenses_branch_id ON softivite.expenses USING btree (branch_id);


--
-- Name: idx_expenses_category_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expenses_category_id ON softivite.expenses USING btree (category_id);


--
-- Name: idx_expenses_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expenses_created_at ON softivite.expenses USING btree (created_at);


--
-- Name: idx_expenses_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expenses_deleted_at ON softivite.expenses USING btree (deleted_at);


--
-- Name: idx_expenses_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expenses_updated_at ON softivite.expenses USING btree (updated_at);


--
-- Name: idx_expenses_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_expenses_version ON softivite.expenses USING btree (version);


--
-- Name: idx_external_systems_client_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_external_systems_client_id ON softivite.external_systems USING btree (client_id);


--
-- Name: idx_external_systems_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_external_systems_created_at ON softivite.external_systems USING btree (created_at);


--
-- Name: idx_external_systems_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_external_systems_deleted_at ON softivite.external_systems USING btree (deleted_at);


--
-- Name: idx_external_systems_developer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_external_systems_developer_id ON softivite.external_systems USING btree (developer_id);


--
-- Name: idx_external_systems_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_external_systems_updated_at ON softivite.external_systems USING btree (updated_at);


--
-- Name: idx_external_systems_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_external_systems_version ON softivite.external_systems USING btree (version);


--
-- Name: idx_gift_cards_code; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_gift_cards_code ON softivite.gift_cards USING btree (code);


--
-- Name: idx_gift_cards_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_gift_cards_created_at ON softivite.gift_cards USING btree (created_at);


--
-- Name: idx_gift_cards_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_gift_cards_deleted_at ON softivite.gift_cards USING btree (deleted_at);


--
-- Name: idx_gift_cards_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_gift_cards_updated_at ON softivite.gift_cards USING btree (updated_at);


--
-- Name: idx_gift_cards_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_gift_cards_version ON softivite.gift_cards USING btree (version);


--
-- Name: idx_journal_entries_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_journal_entries_created_at ON softivite.journal_entries USING btree (created_at);


--
-- Name: idx_journal_entries_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_journal_entries_deleted_at ON softivite.journal_entries USING btree (deleted_at);


--
-- Name: idx_journal_entries_reference_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_journal_entries_reference_id ON softivite.journal_entries USING btree (reference_id);


--
-- Name: idx_journal_entries_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_journal_entries_updated_at ON softivite.journal_entries USING btree (updated_at);


--
-- Name: idx_journal_entries_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_journal_entries_version ON softivite.journal_entries USING btree (version);


--
-- Name: idx_kds_tickets_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_kds_tickets_branch_id ON softivite.kds_tickets USING btree (branch_id);


--
-- Name: idx_kds_tickets_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_kds_tickets_created_at ON softivite.kds_tickets USING btree (created_at);


--
-- Name: idx_kds_tickets_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_kds_tickets_deleted_at ON softivite.kds_tickets USING btree (deleted_at);


--
-- Name: idx_kds_tickets_order_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_kds_tickets_order_id ON softivite.kds_tickets USING btree (order_id);


--
-- Name: idx_kds_tickets_status; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_kds_tickets_status ON softivite.kds_tickets USING btree (status);


--
-- Name: idx_kds_tickets_table_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_kds_tickets_table_id ON softivite.kds_tickets USING btree (table_id);


--
-- Name: idx_kds_tickets_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_kds_tickets_updated_at ON softivite.kds_tickets USING btree (updated_at);


--
-- Name: idx_kds_tickets_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_kds_tickets_version ON softivite.kds_tickets USING btree (version);


--
-- Name: idx_leave_requests_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_leave_requests_created_at ON softivite.leave_requests USING btree (created_at);


--
-- Name: idx_leave_requests_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_leave_requests_deleted_at ON softivite.leave_requests USING btree (deleted_at);


--
-- Name: idx_leave_requests_staff_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_leave_requests_staff_id ON softivite.leave_requests USING btree (staff_id);


--
-- Name: idx_leave_requests_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_leave_requests_updated_at ON softivite.leave_requests USING btree (updated_at);


--
-- Name: idx_leave_requests_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_leave_requests_version ON softivite.leave_requests USING btree (version);


--
-- Name: idx_ledger_accounts_code; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_ledger_accounts_code ON softivite.ledger_accounts USING btree (code);


--
-- Name: idx_ledger_accounts_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_accounts_created_at ON softivite.ledger_accounts USING btree (created_at);


--
-- Name: idx_ledger_accounts_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_accounts_deleted_at ON softivite.ledger_accounts USING btree (deleted_at);


--
-- Name: idx_ledger_accounts_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_accounts_updated_at ON softivite.ledger_accounts USING btree (updated_at);


--
-- Name: idx_ledger_accounts_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_accounts_version ON softivite.ledger_accounts USING btree (version);


--
-- Name: idx_ledger_lines_account_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_lines_account_id ON softivite.ledger_lines USING btree (account_id);


--
-- Name: idx_ledger_lines_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_lines_created_at ON softivite.ledger_lines USING btree (created_at);


--
-- Name: idx_ledger_lines_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_lines_deleted_at ON softivite.ledger_lines USING btree (deleted_at);


--
-- Name: idx_ledger_lines_journal_entry_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_lines_journal_entry_id ON softivite.ledger_lines USING btree (journal_entry_id);


--
-- Name: idx_ledger_lines_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_lines_updated_at ON softivite.ledger_lines USING btree (updated_at);


--
-- Name: idx_ledger_lines_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_ledger_lines_version ON softivite.ledger_lines USING btree (version);


--
-- Name: idx_loyalty_transactions_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_loyalty_transactions_created_at ON softivite.loyalty_transactions USING btree (created_at);


--
-- Name: idx_loyalty_transactions_customer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_loyalty_transactions_customer_id ON softivite.loyalty_transactions USING btree (customer_id);


--
-- Name: idx_loyalty_transactions_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_loyalty_transactions_deleted_at ON softivite.loyalty_transactions USING btree (deleted_at);


--
-- Name: idx_loyalty_transactions_tenant_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_loyalty_transactions_tenant_id ON softivite.loyalty_transactions USING btree (tenant_id);


--
-- Name: idx_loyalty_transactions_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_loyalty_transactions_updated_at ON softivite.loyalty_transactions USING btree (updated_at);


--
-- Name: idx_loyalty_transactions_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_loyalty_transactions_version ON softivite.loyalty_transactions USING btree (version);


--
-- Name: idx_marketing_campaigns_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_marketing_campaigns_created_at ON softivite.marketing_campaigns USING btree (created_at);


--
-- Name: idx_marketing_campaigns_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_marketing_campaigns_deleted_at ON softivite.marketing_campaigns USING btree (deleted_at);


--
-- Name: idx_marketing_campaigns_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_marketing_campaigns_updated_at ON softivite.marketing_campaigns USING btree (updated_at);


--
-- Name: idx_marketing_campaigns_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_marketing_campaigns_version ON softivite.marketing_campaigns USING btree (version);


--
-- Name: idx_newsletter_subscriptions_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_newsletter_subscriptions_created_at ON softivite.newsletter_subscriptions USING btree (created_at);


--
-- Name: idx_newsletter_subscriptions_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_newsletter_subscriptions_deleted_at ON softivite.newsletter_subscriptions USING btree (deleted_at);


--
-- Name: idx_newsletter_subscriptions_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_newsletter_subscriptions_updated_at ON softivite.newsletter_subscriptions USING btree (updated_at);


--
-- Name: idx_newsletter_subscriptions_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_newsletter_subscriptions_version ON softivite.newsletter_subscriptions USING btree (version);


--
-- Name: idx_notification_settings_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notification_settings_created_at ON softivite.notification_settings USING btree (created_at);


--
-- Name: idx_notification_settings_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notification_settings_deleted_at ON softivite.notification_settings USING btree (deleted_at);


--
-- Name: idx_notification_settings_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notification_settings_updated_at ON softivite.notification_settings USING btree (updated_at);


--
-- Name: idx_notification_settings_user_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_notification_settings_user_id ON softivite.notification_settings USING btree (user_id);


--
-- Name: idx_notification_settings_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notification_settings_version ON softivite.notification_settings USING btree (version);


--
-- Name: idx_notifications_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notifications_created_at ON softivite.notifications USING btree (created_at);


--
-- Name: idx_notifications_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notifications_deleted_at ON softivite.notifications USING btree (deleted_at);


--
-- Name: idx_notifications_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notifications_updated_at ON softivite.notifications USING btree (updated_at);


--
-- Name: idx_notifications_user_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notifications_user_id ON softivite.notifications USING btree (user_id);


--
-- Name: idx_notifications_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_notifications_version ON softivite.notifications USING btree (version);


--
-- Name: idx_order_items_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_order_items_created_at ON softivite.order_items USING btree (created_at);


--
-- Name: idx_order_items_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_order_items_deleted_at ON softivite.order_items USING btree (deleted_at);


--
-- Name: idx_order_items_order_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_order_items_order_id ON softivite.order_items USING btree (order_id);


--
-- Name: idx_order_items_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_order_items_product_id ON softivite.order_items USING btree (product_id);


--
-- Name: idx_order_items_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_order_items_updated_at ON softivite.order_items USING btree (updated_at);


--
-- Name: idx_order_items_variant_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_order_items_variant_id ON softivite.order_items USING btree (variant_id);


--
-- Name: idx_order_items_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_order_items_version ON softivite.order_items USING btree (version);


--
-- Name: idx_orders_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_orders_branch_id ON softivite.orders USING btree (branch_id);


--
-- Name: idx_orders_cashier_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_orders_cashier_id ON softivite.orders USING btree (cashier_id);


--
-- Name: idx_orders_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_orders_created_at ON softivite.orders USING btree (created_at);


--
-- Name: idx_orders_customer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_orders_customer_id ON softivite.orders USING btree (customer_id);


--
-- Name: idx_orders_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_orders_deleted_at ON softivite.orders USING btree (deleted_at);


--
-- Name: idx_orders_order_number; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_orders_order_number ON softivite.orders USING btree (order_number);


--
-- Name: idx_orders_receipt_token; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_orders_receipt_token ON softivite.orders USING btree (receipt_token);


--
-- Name: idx_orders_status; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_orders_status ON softivite.orders USING btree (status);


--
-- Name: idx_orders_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_orders_updated_at ON softivite.orders USING btree (updated_at);


--
-- Name: idx_orders_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_orders_version ON softivite.orders USING btree (version);


--
-- Name: idx_payment_methods_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payment_methods_created_at ON softivite.payment_methods USING btree (created_at);


--
-- Name: idx_payment_methods_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payment_methods_deleted_at ON softivite.payment_methods USING btree (deleted_at);


--
-- Name: idx_payment_methods_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payment_methods_updated_at ON softivite.payment_methods USING btree (updated_at);


--
-- Name: idx_payment_methods_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payment_methods_version ON softivite.payment_methods USING btree (version);


--
-- Name: idx_payments_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payments_created_at ON softivite.payments USING btree (created_at);


--
-- Name: idx_payments_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payments_deleted_at ON softivite.payments USING btree (deleted_at);


--
-- Name: idx_payments_order_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payments_order_id ON softivite.payments USING btree (order_id);


--
-- Name: idx_payments_payment_method_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payments_payment_method_id ON softivite.payments USING btree (payment_method_id);


--
-- Name: idx_payments_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payments_updated_at ON softivite.payments USING btree (updated_at);


--
-- Name: idx_payments_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payments_version ON softivite.payments USING btree (version);


--
-- Name: idx_payroll_periods_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_periods_created_at ON softivite.payroll_periods USING btree (created_at);


--
-- Name: idx_payroll_periods_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_periods_deleted_at ON softivite.payroll_periods USING btree (deleted_at);


--
-- Name: idx_payroll_periods_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_periods_updated_at ON softivite.payroll_periods USING btree (updated_at);


--
-- Name: idx_payroll_periods_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_periods_version ON softivite.payroll_periods USING btree (version);


--
-- Name: idx_payroll_records_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_records_created_at ON softivite.payroll_records USING btree (created_at);


--
-- Name: idx_payroll_records_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_records_deleted_at ON softivite.payroll_records USING btree (deleted_at);


--
-- Name: idx_payroll_records_is_paid; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_records_is_paid ON softivite.payroll_records USING btree (is_paid);


--
-- Name: idx_payroll_records_period_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_records_period_id ON softivite.payroll_records USING btree (period_id);


--
-- Name: idx_payroll_records_staff_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_records_staff_id ON softivite.payroll_records USING btree (staff_id);


--
-- Name: idx_payroll_records_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_records_updated_at ON softivite.payroll_records USING btree (updated_at);


--
-- Name: idx_payroll_records_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_payroll_records_version ON softivite.payroll_records USING btree (version);


--
-- Name: idx_print_jobs_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_print_jobs_branch_id ON softivite.print_jobs USING btree (branch_id);


--
-- Name: idx_print_jobs_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_print_jobs_created_at ON softivite.print_jobs USING btree (created_at);


--
-- Name: idx_print_jobs_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_print_jobs_deleted_at ON softivite.print_jobs USING btree (deleted_at);


--
-- Name: idx_print_jobs_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_print_jobs_updated_at ON softivite.print_jobs USING btree (updated_at);


--
-- Name: idx_print_jobs_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_print_jobs_version ON softivite.print_jobs USING btree (version);


--
-- Name: idx_product_components_component_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_components_component_product_id ON softivite.product_components USING btree (component_product_id);


--
-- Name: idx_product_components_composite_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_components_composite_product_id ON softivite.product_components USING btree (composite_product_id);


--
-- Name: idx_product_components_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_components_created_at ON softivite.product_components USING btree (created_at);


--
-- Name: idx_product_components_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_components_deleted_at ON softivite.product_components USING btree (deleted_at);


--
-- Name: idx_product_components_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_components_updated_at ON softivite.product_components USING btree (updated_at);


--
-- Name: idx_product_components_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_components_version ON softivite.product_components USING btree (version);


--
-- Name: idx_product_histories_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_histories_created_at ON softivite.product_histories USING btree (created_at);


--
-- Name: idx_product_histories_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_histories_deleted_at ON softivite.product_histories USING btree (deleted_at);


--
-- Name: idx_product_histories_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_histories_product_id ON softivite.product_histories USING btree (product_id);


--
-- Name: idx_product_histories_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_histories_updated_at ON softivite.product_histories USING btree (updated_at);


--
-- Name: idx_product_histories_user_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_histories_user_id ON softivite.product_histories USING btree (user_id);


--
-- Name: idx_product_histories_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_histories_version ON softivite.product_histories USING btree (version);


--
-- Name: idx_product_image_galleries_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_image_galleries_created_at ON softivite.product_image_galleries USING btree (created_at);


--
-- Name: idx_product_image_galleries_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_image_galleries_deleted_at ON softivite.product_image_galleries USING btree (deleted_at);


--
-- Name: idx_product_image_galleries_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_image_galleries_product_id ON softivite.product_image_galleries USING btree (product_id);


--
-- Name: idx_product_image_galleries_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_image_galleries_updated_at ON softivite.product_image_galleries USING btree (updated_at);


--
-- Name: idx_product_image_galleries_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_image_galleries_version ON softivite.product_image_galleries USING btree (version);


--
-- Name: idx_product_reviews_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_reviews_created_at ON softivite.product_reviews USING btree (created_at);


--
-- Name: idx_product_reviews_customer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_reviews_customer_id ON softivite.product_reviews USING btree (customer_id);


--
-- Name: idx_product_reviews_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_reviews_deleted_at ON softivite.product_reviews USING btree (deleted_at);


--
-- Name: idx_product_reviews_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_reviews_product_id ON softivite.product_reviews USING btree (product_id);


--
-- Name: idx_product_reviews_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_reviews_updated_at ON softivite.product_reviews USING btree (updated_at);


--
-- Name: idx_product_reviews_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_reviews_version ON softivite.product_reviews USING btree (version);


--
-- Name: idx_product_variants_barcode; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_variants_barcode ON softivite.product_variants USING btree (barcode);


--
-- Name: idx_product_variants_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_variants_created_at ON softivite.product_variants USING btree (created_at);


--
-- Name: idx_product_variants_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_variants_deleted_at ON softivite.product_variants USING btree (deleted_at);


--
-- Name: idx_product_variants_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_variants_product_id ON softivite.product_variants USING btree (product_id);


--
-- Name: idx_product_variants_sku; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_product_variants_sku ON softivite.product_variants USING btree (sku);


--
-- Name: idx_product_variants_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_variants_updated_at ON softivite.product_variants USING btree (updated_at);


--
-- Name: idx_product_variants_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_product_variants_version ON softivite.product_variants USING btree (version);


--
-- Name: idx_products_barcode; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_barcode ON softivite.products USING btree (barcode);


--
-- Name: idx_products_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_branch_id ON softivite.products USING btree (branch_id);


--
-- Name: idx_products_category_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_category_id ON softivite.products USING btree (category_id);


--
-- Name: idx_products_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_created_at ON softivite.products USING btree (created_at);


--
-- Name: idx_products_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_deleted_at ON softivite.products USING btree (deleted_at);


--
-- Name: idx_products_has_variants; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_has_variants ON softivite.products USING btree (has_variants);


--
-- Name: idx_products_is_active; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_is_active ON softivite.products USING btree (is_active);


--
-- Name: idx_products_is_online; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_is_online ON softivite.products USING btree (is_online);


--
-- Name: idx_products_supplier_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_supplier_id ON softivite.products USING btree (supplier_id);


--
-- Name: idx_products_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_updated_at ON softivite.products USING btree (updated_at);


--
-- Name: idx_products_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_products_version ON softivite.products USING btree (version);


--
-- Name: idx_promotions_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_promotions_deleted_at ON softivite.promotions USING btree (deleted_at);


--
-- Name: idx_purchase_order_items_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_order_items_created_at ON softivite.purchase_order_items USING btree (created_at);


--
-- Name: idx_purchase_order_items_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_order_items_deleted_at ON softivite.purchase_order_items USING btree (deleted_at);


--
-- Name: idx_purchase_order_items_po_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_order_items_po_id ON softivite.purchase_order_items USING btree (po_id);


--
-- Name: idx_purchase_order_items_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_order_items_product_id ON softivite.purchase_order_items USING btree (product_id);


--
-- Name: idx_purchase_order_items_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_order_items_updated_at ON softivite.purchase_order_items USING btree (updated_at);


--
-- Name: idx_purchase_order_items_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_order_items_version ON softivite.purchase_order_items USING btree (version);


--
-- Name: idx_purchase_orders_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_orders_branch_id ON softivite.purchase_orders USING btree (branch_id);


--
-- Name: idx_purchase_orders_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_orders_created_at ON softivite.purchase_orders USING btree (created_at);


--
-- Name: idx_purchase_orders_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_orders_deleted_at ON softivite.purchase_orders USING btree (deleted_at);


--
-- Name: idx_purchase_orders_po_number; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_purchase_orders_po_number ON softivite.purchase_orders USING btree (po_number);


--
-- Name: idx_purchase_orders_status; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_orders_status ON softivite.purchase_orders USING btree (status);


--
-- Name: idx_purchase_orders_supplier_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_orders_supplier_id ON softivite.purchase_orders USING btree (supplier_id);


--
-- Name: idx_purchase_orders_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_orders_updated_at ON softivite.purchase_orders USING btree (updated_at);


--
-- Name: idx_purchase_orders_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_purchase_orders_version ON softivite.purchase_orders USING btree (version);


--
-- Name: idx_quotation_items_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotation_items_created_at ON softivite.quotation_items USING btree (created_at);


--
-- Name: idx_quotation_items_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotation_items_deleted_at ON softivite.quotation_items USING btree (deleted_at);


--
-- Name: idx_quotation_items_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotation_items_product_id ON softivite.quotation_items USING btree (product_id);


--
-- Name: idx_quotation_items_quotation_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotation_items_quotation_id ON softivite.quotation_items USING btree (quotation_id);


--
-- Name: idx_quotation_items_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotation_items_updated_at ON softivite.quotation_items USING btree (updated_at);


--
-- Name: idx_quotation_items_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotation_items_version ON softivite.quotation_items USING btree (version);


--
-- Name: idx_quotations_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotations_branch_id ON softivite.quotations USING btree (branch_id);


--
-- Name: idx_quotations_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotations_created_at ON softivite.quotations USING btree (created_at);


--
-- Name: idx_quotations_customer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotations_customer_id ON softivite.quotations USING btree (customer_id);


--
-- Name: idx_quotations_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotations_deleted_at ON softivite.quotations USING btree (deleted_at);


--
-- Name: idx_quotations_quote_number; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_quotations_quote_number ON softivite.quotations USING btree (quote_number);


--
-- Name: idx_quotations_status; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotations_status ON softivite.quotations USING btree (status);


--
-- Name: idx_quotations_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotations_updated_at ON softivite.quotations USING btree (updated_at);


--
-- Name: idx_quotations_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_quotations_version ON softivite.quotations USING btree (version);


--
-- Name: idx_return_items_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_return_items_created_at ON softivite.return_items USING btree (created_at);


--
-- Name: idx_return_items_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_return_items_deleted_at ON softivite.return_items USING btree (deleted_at);


--
-- Name: idx_return_items_return_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_return_items_return_id ON softivite.return_items USING btree (return_id);


--
-- Name: idx_return_items_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_return_items_updated_at ON softivite.return_items USING btree (updated_at);


--
-- Name: idx_return_items_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_return_items_version ON softivite.return_items USING btree (version);


--
-- Name: idx_returns_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_returns_branch_id ON softivite.returns USING btree (branch_id);


--
-- Name: idx_returns_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_returns_created_at ON softivite.returns USING btree (created_at);


--
-- Name: idx_returns_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_returns_deleted_at ON softivite.returns USING btree (deleted_at);


--
-- Name: idx_returns_order_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_returns_order_id ON softivite.returns USING btree (order_id);


--
-- Name: idx_returns_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_returns_updated_at ON softivite.returns USING btree (updated_at);


--
-- Name: idx_returns_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_returns_version ON softivite.returns USING btree (version);


--
-- Name: idx_service_categories_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_categories_created_at ON softivite.service_categories USING btree (created_at);


--
-- Name: idx_service_categories_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_categories_deleted_at ON softivite.service_categories USING btree (deleted_at);


--
-- Name: idx_service_categories_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_categories_updated_at ON softivite.service_categories USING btree (updated_at);


--
-- Name: idx_service_categories_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_categories_version ON softivite.service_categories USING btree (version);


--
-- Name: idx_service_commission_records_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_records_created_at ON softivite.service_commission_records USING btree (created_at);


--
-- Name: idx_service_commission_records_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_records_deleted_at ON softivite.service_commission_records USING btree (deleted_at);


--
-- Name: idx_service_commission_records_is_paid; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_records_is_paid ON softivite.service_commission_records USING btree (is_paid);


--
-- Name: idx_service_commission_records_order_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_records_order_id ON softivite.service_commission_records USING btree (order_id);


--
-- Name: idx_service_commission_records_staff_member_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_records_staff_member_id ON softivite.service_commission_records USING btree (staff_member_id);


--
-- Name: idx_service_commission_records_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_records_updated_at ON softivite.service_commission_records USING btree (updated_at);


--
-- Name: idx_service_commission_records_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_records_version ON softivite.service_commission_records USING btree (version);


--
-- Name: idx_service_commission_rules_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_rules_created_at ON softivite.service_commission_rules USING btree (created_at);


--
-- Name: idx_service_commission_rules_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_rules_deleted_at ON softivite.service_commission_rules USING btree (deleted_at);


--
-- Name: idx_service_commission_rules_staff_member_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_rules_staff_member_id ON softivite.service_commission_rules USING btree (staff_member_id);


--
-- Name: idx_service_commission_rules_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_rules_updated_at ON softivite.service_commission_rules USING btree (updated_at);


--
-- Name: idx_service_commission_rules_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_service_commission_rules_version ON softivite.service_commission_rules USING btree (version);


--
-- Name: idx_services_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_services_branch_id ON softivite.services USING btree (branch_id);


--
-- Name: idx_services_category_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_services_category_id ON softivite.services USING btree (category_id);


--
-- Name: idx_services_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_services_created_at ON softivite.services USING btree (created_at);


--
-- Name: idx_services_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_services_deleted_at ON softivite.services USING btree (deleted_at);


--
-- Name: idx_services_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_services_updated_at ON softivite.services USING btree (updated_at);


--
-- Name: idx_services_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_services_version ON softivite.services USING btree (version);


--
-- Name: idx_shift_swap_requests_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shift_swap_requests_created_at ON softivite.shift_swap_requests USING btree (created_at);


--
-- Name: idx_shift_swap_requests_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shift_swap_requests_deleted_at ON softivite.shift_swap_requests USING btree (deleted_at);


--
-- Name: idx_shift_swap_requests_original_shift_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shift_swap_requests_original_shift_id ON softivite.shift_swap_requests USING btree (original_shift_id);


--
-- Name: idx_shift_swap_requests_requesting_staff_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shift_swap_requests_requesting_staff_id ON softivite.shift_swap_requests USING btree (requesting_staff_id);


--
-- Name: idx_shift_swap_requests_target_staff_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shift_swap_requests_target_staff_id ON softivite.shift_swap_requests USING btree (target_staff_id);


--
-- Name: idx_shift_swap_requests_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shift_swap_requests_updated_at ON softivite.shift_swap_requests USING btree (updated_at);


--
-- Name: idx_shift_swap_requests_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shift_swap_requests_version ON softivite.shift_swap_requests USING btree (version);


--
-- Name: idx_shifts_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shifts_branch_id ON softivite.shifts USING btree (branch_id);


--
-- Name: idx_shifts_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shifts_created_at ON softivite.shifts USING btree (created_at);


--
-- Name: idx_shifts_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shifts_deleted_at ON softivite.shifts USING btree (deleted_at);


--
-- Name: idx_shifts_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shifts_updated_at ON softivite.shifts USING btree (updated_at);


--
-- Name: idx_shifts_user_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shifts_user_id ON softivite.shifts USING btree (user_id);


--
-- Name: idx_shifts_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_shifts_version ON softivite.shifts USING btree (version);


--
-- Name: idx_split_bill_groups_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_split_bill_groups_branch_id ON softivite.split_bill_groups USING btree (branch_id);


--
-- Name: idx_split_bill_groups_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_split_bill_groups_created_at ON softivite.split_bill_groups USING btree (created_at);


--
-- Name: idx_split_bill_groups_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_split_bill_groups_deleted_at ON softivite.split_bill_groups USING btree (deleted_at);


--
-- Name: idx_split_bill_groups_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_split_bill_groups_updated_at ON softivite.split_bill_groups USING btree (updated_at);


--
-- Name: idx_split_bill_groups_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_split_bill_groups_version ON softivite.split_bill_groups USING btree (version);


--
-- Name: idx_staff_achievements_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_staff_achievements_created_at ON softivite.staff_achievements USING btree (created_at);


--
-- Name: idx_staff_achievements_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_staff_achievements_deleted_at ON softivite.staff_achievements USING btree (deleted_at);


--
-- Name: idx_staff_achievements_staff_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_staff_achievements_staff_id ON softivite.staff_achievements USING btree (staff_id);


--
-- Name: idx_staff_achievements_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_staff_achievements_updated_at ON softivite.staff_achievements USING btree (updated_at);


--
-- Name: idx_staff_achievements_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_staff_achievements_version ON softivite.staff_achievements USING btree (version);


--
-- Name: idx_stock_batches_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_batches_branch_id ON softivite.stock_batches USING btree (branch_id);


--
-- Name: idx_stock_batches_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_batches_created_at ON softivite.stock_batches USING btree (created_at);


--
-- Name: idx_stock_batches_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_batches_deleted_at ON softivite.stock_batches USING btree (deleted_at);


--
-- Name: idx_stock_batches_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_batches_product_id ON softivite.stock_batches USING btree (product_id);


--
-- Name: idx_stock_batches_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_batches_updated_at ON softivite.stock_batches USING btree (updated_at);


--
-- Name: idx_stock_batches_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_batches_version ON softivite.stock_batches USING btree (version);


--
-- Name: idx_stock_movements_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_movements_branch_id ON softivite.stock_movements USING btree (branch_id);


--
-- Name: idx_stock_movements_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_movements_created_at ON softivite.stock_movements USING btree (created_at);


--
-- Name: idx_stock_movements_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_movements_deleted_at ON softivite.stock_movements USING btree (deleted_at);


--
-- Name: idx_stock_movements_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_movements_product_id ON softivite.stock_movements USING btree (product_id);


--
-- Name: idx_stock_movements_tenant_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_movements_tenant_id ON softivite.stock_movements USING btree (tenant_id);


--
-- Name: idx_stock_movements_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_movements_updated_at ON softivite.stock_movements USING btree (updated_at);


--
-- Name: idx_stock_movements_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_movements_version ON softivite.stock_movements USING btree (version);


--
-- Name: idx_stock_transfer_items_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfer_items_created_at ON softivite.stock_transfer_items USING btree (created_at);


--
-- Name: idx_stock_transfer_items_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfer_items_deleted_at ON softivite.stock_transfer_items USING btree (deleted_at);


--
-- Name: idx_stock_transfer_items_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfer_items_product_id ON softivite.stock_transfer_items USING btree (product_id);


--
-- Name: idx_stock_transfer_items_transfer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfer_items_transfer_id ON softivite.stock_transfer_items USING btree (transfer_id);


--
-- Name: idx_stock_transfer_items_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfer_items_updated_at ON softivite.stock_transfer_items USING btree (updated_at);


--
-- Name: idx_stock_transfer_items_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfer_items_version ON softivite.stock_transfer_items USING btree (version);


--
-- Name: idx_stock_transfers_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfers_created_at ON softivite.stock_transfers USING btree (created_at);


--
-- Name: idx_stock_transfers_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfers_deleted_at ON softivite.stock_transfers USING btree (deleted_at);


--
-- Name: idx_stock_transfers_from_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfers_from_branch_id ON softivite.stock_transfers USING btree (from_branch_id);


--
-- Name: idx_stock_transfers_reference_no; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_stock_transfers_reference_no ON softivite.stock_transfers USING btree (reference_no);


--
-- Name: idx_stock_transfers_status; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfers_status ON softivite.stock_transfers USING btree (status);


--
-- Name: idx_stock_transfers_to_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfers_to_branch_id ON softivite.stock_transfers USING btree (to_branch_id);


--
-- Name: idx_stock_transfers_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfers_updated_at ON softivite.stock_transfers USING btree (updated_at);


--
-- Name: idx_stock_transfers_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stock_transfers_version ON softivite.stock_transfers USING btree (version);


--
-- Name: idx_stocktake_entries_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_entries_created_at ON softivite.stocktake_entries USING btree (created_at);


--
-- Name: idx_stocktake_entries_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_entries_deleted_at ON softivite.stocktake_entries USING btree (deleted_at);


--
-- Name: idx_stocktake_entries_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_entries_product_id ON softivite.stocktake_entries USING btree (product_id);


--
-- Name: idx_stocktake_entries_session_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_entries_session_id ON softivite.stocktake_entries USING btree (session_id);


--
-- Name: idx_stocktake_entries_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_entries_updated_at ON softivite.stocktake_entries USING btree (updated_at);


--
-- Name: idx_stocktake_entries_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_entries_version ON softivite.stocktake_entries USING btree (version);


--
-- Name: idx_stocktake_sessions_access_token; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_stocktake_sessions_access_token ON softivite.stocktake_sessions USING btree (access_token);


--
-- Name: idx_stocktake_sessions_branch_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_sessions_branch_id ON softivite.stocktake_sessions USING btree (branch_id);


--
-- Name: idx_stocktake_sessions_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_sessions_created_at ON softivite.stocktake_sessions USING btree (created_at);


--
-- Name: idx_stocktake_sessions_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_sessions_deleted_at ON softivite.stocktake_sessions USING btree (deleted_at);


--
-- Name: idx_stocktake_sessions_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_sessions_updated_at ON softivite.stocktake_sessions USING btree (updated_at);


--
-- Name: idx_stocktake_sessions_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_stocktake_sessions_version ON softivite.stocktake_sessions USING btree (version);


--
-- Name: idx_storefront_settings_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_storefront_settings_created_at ON softivite.storefront_settings USING btree (created_at);


--
-- Name: idx_storefront_settings_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_storefront_settings_deleted_at ON softivite.storefront_settings USING btree (deleted_at);


--
-- Name: idx_storefront_settings_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_storefront_settings_updated_at ON softivite.storefront_settings USING btree (updated_at);


--
-- Name: idx_storefront_settings_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_storefront_settings_version ON softivite.storefront_settings USING btree (version);


--
-- Name: idx_supplier_ledger_entries_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_created_at ON softivite.supplier_ledger_entries USING btree (created_at);


--
-- Name: idx_supplier_ledger_entries_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_deleted_at ON softivite.supplier_ledger_entries USING btree (deleted_at);


--
-- Name: idx_supplier_ledger_entries_supplier_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_supplier_id ON softivite.supplier_ledger_entries USING btree (supplier_id);


--
-- Name: idx_supplier_ledger_entries_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_updated_at ON softivite.supplier_ledger_entries USING btree (updated_at);


--
-- Name: idx_supplier_ledger_entries_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_version ON softivite.supplier_ledger_entries USING btree (version);


--
-- Name: idx_supplier_product; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_supplier_product ON softivite.supplier_products USING btree (supplier_id, product_id);


--
-- Name: idx_supplier_products_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_products_created_at ON softivite.supplier_products USING btree (created_at);


--
-- Name: idx_supplier_products_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_products_deleted_at ON softivite.supplier_products USING btree (deleted_at);


--
-- Name: idx_supplier_products_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_products_product_id ON softivite.supplier_products USING btree (product_id);


--
-- Name: idx_supplier_products_supplier_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_products_supplier_id ON softivite.supplier_products USING btree (supplier_id);


--
-- Name: idx_supplier_products_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_products_updated_at ON softivite.supplier_products USING btree (updated_at);


--
-- Name: idx_supplier_products_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_products_version ON softivite.supplier_products USING btree (version);


--
-- Name: idx_supplier_profiles_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_profiles_created_at ON softivite.supplier_profiles USING btree (created_at);


--
-- Name: idx_supplier_profiles_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_profiles_deleted_at ON softivite.supplier_profiles USING btree (deleted_at);


--
-- Name: idx_supplier_profiles_supplier_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_profiles_supplier_id ON softivite.supplier_profiles USING btree (supplier_id);


--
-- Name: idx_supplier_profiles_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_profiles_updated_at ON softivite.supplier_profiles USING btree (updated_at);


--
-- Name: idx_supplier_profiles_user_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_profiles_user_id ON softivite.supplier_profiles USING btree (user_id);


--
-- Name: idx_supplier_profiles_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_supplier_profiles_version ON softivite.supplier_profiles USING btree (version);


--
-- Name: idx_suppliers_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_suppliers_created_at ON softivite.suppliers USING btree (created_at);


--
-- Name: idx_suppliers_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_suppliers_deleted_at ON softivite.suppliers USING btree (deleted_at);


--
-- Name: idx_suppliers_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_suppliers_updated_at ON softivite.suppliers USING btree (updated_at);


--
-- Name: idx_suppliers_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_suppliers_version ON softivite.suppliers USING btree (version);


--
-- Name: idx_tax_configurations_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_tax_configurations_created_at ON softivite.tax_configurations USING btree (created_at);


--
-- Name: idx_tax_configurations_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_tax_configurations_deleted_at ON softivite.tax_configurations USING btree (deleted_at);


--
-- Name: idx_tax_configurations_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_tax_configurations_updated_at ON softivite.tax_configurations USING btree (updated_at);


--
-- Name: idx_tax_configurations_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_tax_configurations_version ON softivite.tax_configurations USING btree (version);


--
-- Name: idx_tenant_branch_sku; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_branch_sku ON softivite.products USING btree (branch_id, sku);


--
-- Name: idx_tenant_integrations_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_tenant_integrations_created_at ON softivite.tenant_integrations USING btree (created_at);


--
-- Name: idx_tenant_integrations_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_tenant_integrations_deleted_at ON softivite.tenant_integrations USING btree (deleted_at);


--
-- Name: idx_tenant_integrations_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_tenant_integrations_updated_at ON softivite.tenant_integrations USING btree (updated_at);


--
-- Name: idx_tenant_integrations_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_tenant_integrations_version ON softivite.tenant_integrations USING btree (version);


--
-- Name: idx_tenant_provider; Type: INDEX; Schema: softivite; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_provider ON softivite.tenant_integrations USING btree (provider);


--
-- Name: idx_webhook_endpoints_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_endpoints_created_at ON softivite.webhook_endpoints USING btree (created_at);


--
-- Name: idx_webhook_endpoints_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_endpoints_deleted_at ON softivite.webhook_endpoints USING btree (deleted_at);


--
-- Name: idx_webhook_endpoints_external_system_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_endpoints_external_system_id ON softivite.webhook_endpoints USING btree (external_system_id);


--
-- Name: idx_webhook_endpoints_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_endpoints_updated_at ON softivite.webhook_endpoints USING btree (updated_at);


--
-- Name: idx_webhook_endpoints_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_endpoints_version ON softivite.webhook_endpoints USING btree (version);


--
-- Name: idx_webhook_events_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_events_created_at ON softivite.webhook_events USING btree (created_at);


--
-- Name: idx_webhook_events_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_events_deleted_at ON softivite.webhook_events USING btree (deleted_at);


--
-- Name: idx_webhook_events_endpoint_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_events_endpoint_id ON softivite.webhook_events USING btree (endpoint_id);


--
-- Name: idx_webhook_events_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_events_updated_at ON softivite.webhook_events USING btree (updated_at);


--
-- Name: idx_webhook_events_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_webhook_events_version ON softivite.webhook_events USING btree (version);


--
-- Name: idx_wishlists_created_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_wishlists_created_at ON softivite.wishlists USING btree (created_at);


--
-- Name: idx_wishlists_customer_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_wishlists_customer_id ON softivite.wishlists USING btree (customer_id);


--
-- Name: idx_wishlists_deleted_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_wishlists_deleted_at ON softivite.wishlists USING btree (deleted_at);


--
-- Name: idx_wishlists_product_id; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_wishlists_product_id ON softivite.wishlists USING btree (product_id);


--
-- Name: idx_wishlists_updated_at; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_wishlists_updated_at ON softivite.wishlists USING btree (updated_at);


--
-- Name: idx_wishlists_version; Type: INDEX; Schema: softivite; Owner: -
--

CREATE INDEX idx_wishlists_version ON softivite.wishlists USING btree (version);


--
-- Name: idx_abandoned_carts_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_abandoned_carts_created_at ON thinkce.abandoned_carts USING btree (created_at);


--
-- Name: idx_abandoned_carts_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_abandoned_carts_deleted_at ON thinkce.abandoned_carts USING btree (deleted_at);


--
-- Name: idx_abandoned_carts_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_abandoned_carts_updated_at ON thinkce.abandoned_carts USING btree (updated_at);


--
-- Name: idx_abandoned_carts_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_abandoned_carts_version ON thinkce.abandoned_carts USING btree (version);


--
-- Name: idx_api_keys_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_keys_branch_id ON thinkce.api_keys USING btree (branch_id);


--
-- Name: idx_api_keys_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_keys_created_at ON thinkce.api_keys USING btree (created_at);


--
-- Name: idx_api_keys_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_keys_deleted_at ON thinkce.api_keys USING btree (deleted_at);


--
-- Name: idx_api_keys_key_prefix; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_keys_key_prefix ON thinkce.api_keys USING btree (key_prefix);


--
-- Name: idx_api_keys_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_keys_updated_at ON thinkce.api_keys USING btree (updated_at);


--
-- Name: idx_api_keys_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_keys_version ON thinkce.api_keys USING btree (version);


--
-- Name: idx_api_request_logs_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_request_logs_created_at ON thinkce.api_request_logs USING btree (created_at);


--
-- Name: idx_api_request_logs_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_request_logs_deleted_at ON thinkce.api_request_logs USING btree (deleted_at);


--
-- Name: idx_api_request_logs_endpoint; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_request_logs_endpoint ON thinkce.api_request_logs USING btree (endpoint);


--
-- Name: idx_api_request_logs_status_code; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_request_logs_status_code ON thinkce.api_request_logs USING btree (status_code);


--
-- Name: idx_api_request_logs_tenant_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_request_logs_tenant_id ON thinkce.api_request_logs USING btree (tenant_id);


--
-- Name: idx_api_request_logs_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_request_logs_updated_at ON thinkce.api_request_logs USING btree (updated_at);


--
-- Name: idx_api_request_logs_user_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_request_logs_user_id ON thinkce.api_request_logs USING btree (user_id);


--
-- Name: idx_api_request_logs_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_api_request_logs_version ON thinkce.api_request_logs USING btree (version);


--
-- Name: idx_appointments_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_branch_id ON thinkce.appointments USING btree (branch_id);


--
-- Name: idx_appointments_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_created_at ON thinkce.appointments USING btree (created_at);


--
-- Name: idx_appointments_customer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_customer_id ON thinkce.appointments USING btree (customer_id);


--
-- Name: idx_appointments_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_deleted_at ON thinkce.appointments USING btree (deleted_at);


--
-- Name: idx_appointments_service_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_service_id ON thinkce.appointments USING btree (service_id);


--
-- Name: idx_appointments_staff_member_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_staff_member_id ON thinkce.appointments USING btree (staff_member_id);


--
-- Name: idx_appointments_start_time; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_start_time ON thinkce.appointments USING btree (start_time);


--
-- Name: idx_appointments_status; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_status ON thinkce.appointments USING btree (status);


--
-- Name: idx_appointments_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_updated_at ON thinkce.appointments USING btree (updated_at);


--
-- Name: idx_appointments_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_appointments_version ON thinkce.appointments USING btree (version);


--
-- Name: idx_attendances_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_attendances_branch_id ON thinkce.attendances USING btree (branch_id);


--
-- Name: idx_attendances_clock_in; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_attendances_clock_in ON thinkce.attendances USING btree (clock_in);


--
-- Name: idx_attendances_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_attendances_created_at ON thinkce.attendances USING btree (created_at);


--
-- Name: idx_attendances_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_attendances_deleted_at ON thinkce.attendances USING btree (deleted_at);


--
-- Name: idx_attendances_staff_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_attendances_staff_id ON thinkce.attendances USING btree (staff_id);


--
-- Name: idx_attendances_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_attendances_updated_at ON thinkce.attendances USING btree (updated_at);


--
-- Name: idx_attendances_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_attendances_version ON thinkce.attendances USING btree (version);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_audit_logs_action ON thinkce.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_audit_logs_created_at ON thinkce.audit_logs USING btree (created_at);


--
-- Name: idx_audit_logs_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_audit_logs_deleted_at ON thinkce.audit_logs USING btree (deleted_at);


--
-- Name: idx_audit_logs_tenant_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_audit_logs_tenant_id ON thinkce.audit_logs USING btree (tenant_id);


--
-- Name: idx_audit_logs_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_audit_logs_updated_at ON thinkce.audit_logs USING btree (updated_at);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON thinkce.audit_logs USING btree (user_id);


--
-- Name: idx_audit_logs_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_audit_logs_version ON thinkce.audit_logs USING btree (version);


--
-- Name: idx_branches_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_branches_deleted_at ON thinkce.branches USING btree (deleted_at);


--
-- Name: idx_branches_tenant_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_branches_tenant_id ON thinkce.branches USING btree (tenant_id);


--
-- Name: idx_branches_unique_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_branches_unique_id ON thinkce.branches USING btree (unique_id);


--
-- Name: idx_cash_drawer_sessions_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_branch_id ON thinkce.cash_drawer_sessions USING btree (branch_id);


--
-- Name: idx_cash_drawer_sessions_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_created_at ON thinkce.cash_drawer_sessions USING btree (created_at);


--
-- Name: idx_cash_drawer_sessions_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_deleted_at ON thinkce.cash_drawer_sessions USING btree (deleted_at);


--
-- Name: idx_cash_drawer_sessions_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_updated_at ON thinkce.cash_drawer_sessions USING btree (updated_at);


--
-- Name: idx_cash_drawer_sessions_user_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_user_id ON thinkce.cash_drawer_sessions USING btree (user_id);


--
-- Name: idx_cash_drawer_sessions_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_cash_drawer_sessions_version ON thinkce.cash_drawer_sessions USING btree (version);


--
-- Name: idx_categories_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_categories_created_at ON thinkce.categories USING btree (created_at);


--
-- Name: idx_categories_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_categories_deleted_at ON thinkce.categories USING btree (deleted_at);


--
-- Name: idx_categories_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_categories_updated_at ON thinkce.categories USING btree (updated_at);


--
-- Name: idx_categories_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_categories_version ON thinkce.categories USING btree (version);


--
-- Name: idx_coupons_code; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_coupons_code ON thinkce.coupons USING btree (code);


--
-- Name: idx_coupons_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_coupons_created_at ON thinkce.coupons USING btree (created_at);


--
-- Name: idx_coupons_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_coupons_deleted_at ON thinkce.coupons USING btree (deleted_at);


--
-- Name: idx_coupons_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_coupons_updated_at ON thinkce.coupons USING btree (updated_at);


--
-- Name: idx_coupons_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_coupons_version ON thinkce.coupons USING btree (version);


--
-- Name: idx_crm_settings_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_crm_settings_created_at ON thinkce.crm_settings USING btree (created_at);


--
-- Name: idx_crm_settings_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_crm_settings_deleted_at ON thinkce.crm_settings USING btree (deleted_at);


--
-- Name: idx_crm_settings_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_crm_settings_updated_at ON thinkce.crm_settings USING btree (updated_at);


--
-- Name: idx_crm_settings_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_crm_settings_version ON thinkce.crm_settings USING btree (version);


--
-- Name: idx_customer_feedbacks_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_feedbacks_created_at ON thinkce.customer_feedbacks USING btree (created_at);


--
-- Name: idx_customer_feedbacks_customer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_feedbacks_customer_id ON thinkce.customer_feedbacks USING btree (customer_id);


--
-- Name: idx_customer_feedbacks_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_feedbacks_deleted_at ON thinkce.customer_feedbacks USING btree (deleted_at);


--
-- Name: idx_customer_feedbacks_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_feedbacks_updated_at ON thinkce.customer_feedbacks USING btree (updated_at);


--
-- Name: idx_customer_feedbacks_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_feedbacks_version ON thinkce.customer_feedbacks USING btree (version);


--
-- Name: idx_customer_segments_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_segments_created_at ON thinkce.customer_segments USING btree (created_at);


--
-- Name: idx_customer_segments_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_segments_deleted_at ON thinkce.customer_segments USING btree (deleted_at);


--
-- Name: idx_customer_segments_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_segments_updated_at ON thinkce.customer_segments USING btree (updated_at);


--
-- Name: idx_customer_segments_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_segments_version ON thinkce.customer_segments USING btree (version);


--
-- Name: idx_customer_tiers_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_tiers_created_at ON thinkce.customer_tiers USING btree (created_at);


--
-- Name: idx_customer_tiers_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_tiers_deleted_at ON thinkce.customer_tiers USING btree (deleted_at);


--
-- Name: idx_customer_tiers_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_tiers_updated_at ON thinkce.customer_tiers USING btree (updated_at);


--
-- Name: idx_customer_tiers_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customer_tiers_version ON thinkce.customer_tiers USING btree (version);


--
-- Name: idx_customers_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customers_created_at ON thinkce.customers USING btree (created_at);


--
-- Name: idx_customers_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customers_deleted_at ON thinkce.customers USING btree (deleted_at);


--
-- Name: idx_customers_email; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customers_email ON thinkce.customers USING btree (email);


--
-- Name: idx_customers_name; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customers_name ON thinkce.customers USING btree (name);


--
-- Name: idx_customers_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customers_updated_at ON thinkce.customers USING btree (updated_at);


--
-- Name: idx_customers_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_customers_version ON thinkce.customers USING btree (version);


--
-- Name: idx_delivery_drivers_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_drivers_created_at ON thinkce.delivery_drivers USING btree (created_at);


--
-- Name: idx_delivery_drivers_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_drivers_deleted_at ON thinkce.delivery_drivers USING btree (deleted_at);


--
-- Name: idx_delivery_drivers_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_drivers_updated_at ON thinkce.delivery_drivers USING btree (updated_at);


--
-- Name: idx_delivery_drivers_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_drivers_version ON thinkce.delivery_drivers USING btree (version);


--
-- Name: idx_delivery_orders_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_orders_created_at ON thinkce.delivery_orders USING btree (created_at);


--
-- Name: idx_delivery_orders_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_orders_deleted_at ON thinkce.delivery_orders USING btree (deleted_at);


--
-- Name: idx_delivery_orders_driver_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_orders_driver_id ON thinkce.delivery_orders USING btree (driver_id);


--
-- Name: idx_delivery_orders_order_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_delivery_orders_order_id ON thinkce.delivery_orders USING btree (order_id);


--
-- Name: idx_delivery_orders_tracking_link; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_delivery_orders_tracking_link ON thinkce.delivery_orders USING btree (tracking_link);


--
-- Name: idx_delivery_orders_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_orders_updated_at ON thinkce.delivery_orders USING btree (updated_at);


--
-- Name: idx_delivery_orders_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_delivery_orders_version ON thinkce.delivery_orders USING btree (version);


--
-- Name: idx_dining_tables_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_dining_tables_branch_id ON thinkce.dining_tables USING btree (branch_id);


--
-- Name: idx_dining_tables_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_dining_tables_created_at ON thinkce.dining_tables USING btree (created_at);


--
-- Name: idx_dining_tables_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_dining_tables_deleted_at ON thinkce.dining_tables USING btree (deleted_at);


--
-- Name: idx_dining_tables_status; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_dining_tables_status ON thinkce.dining_tables USING btree (status);


--
-- Name: idx_dining_tables_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_dining_tables_updated_at ON thinkce.dining_tables USING btree (updated_at);


--
-- Name: idx_dining_tables_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_dining_tables_version ON thinkce.dining_tables USING btree (version);


--
-- Name: idx_discount_codes_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_discount_codes_deleted_at ON thinkce.discount_codes USING btree (deleted_at);


--
-- Name: idx_domains_domain; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_domains_domain ON thinkce.domains USING btree (domain);


--
-- Name: idx_domains_tenant_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_domains_tenant_id ON thinkce.domains USING btree (tenant_id);


--
-- Name: idx_expense_categories_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expense_categories_created_at ON thinkce.expense_categories USING btree (created_at);


--
-- Name: idx_expense_categories_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expense_categories_deleted_at ON thinkce.expense_categories USING btree (deleted_at);


--
-- Name: idx_expense_categories_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expense_categories_updated_at ON thinkce.expense_categories USING btree (updated_at);


--
-- Name: idx_expense_categories_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expense_categories_version ON thinkce.expense_categories USING btree (version);


--
-- Name: idx_expenses_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expenses_branch_id ON thinkce.expenses USING btree (branch_id);


--
-- Name: idx_expenses_category_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expenses_category_id ON thinkce.expenses USING btree (category_id);


--
-- Name: idx_expenses_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expenses_created_at ON thinkce.expenses USING btree (created_at);


--
-- Name: idx_expenses_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expenses_deleted_at ON thinkce.expenses USING btree (deleted_at);


--
-- Name: idx_expenses_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expenses_updated_at ON thinkce.expenses USING btree (updated_at);


--
-- Name: idx_expenses_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_expenses_version ON thinkce.expenses USING btree (version);


--
-- Name: idx_external_systems_client_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_external_systems_client_id ON thinkce.external_systems USING btree (client_id);


--
-- Name: idx_external_systems_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_external_systems_created_at ON thinkce.external_systems USING btree (created_at);


--
-- Name: idx_external_systems_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_external_systems_deleted_at ON thinkce.external_systems USING btree (deleted_at);


--
-- Name: idx_external_systems_developer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_external_systems_developer_id ON thinkce.external_systems USING btree (developer_id);


--
-- Name: idx_external_systems_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_external_systems_updated_at ON thinkce.external_systems USING btree (updated_at);


--
-- Name: idx_external_systems_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_external_systems_version ON thinkce.external_systems USING btree (version);


--
-- Name: idx_gift_cards_code; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_gift_cards_code ON thinkce.gift_cards USING btree (code);


--
-- Name: idx_gift_cards_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_gift_cards_created_at ON thinkce.gift_cards USING btree (created_at);


--
-- Name: idx_gift_cards_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_gift_cards_deleted_at ON thinkce.gift_cards USING btree (deleted_at);


--
-- Name: idx_gift_cards_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_gift_cards_updated_at ON thinkce.gift_cards USING btree (updated_at);


--
-- Name: idx_gift_cards_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_gift_cards_version ON thinkce.gift_cards USING btree (version);


--
-- Name: idx_journal_entries_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_journal_entries_created_at ON thinkce.journal_entries USING btree (created_at);


--
-- Name: idx_journal_entries_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_journal_entries_deleted_at ON thinkce.journal_entries USING btree (deleted_at);


--
-- Name: idx_journal_entries_reference_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_journal_entries_reference_id ON thinkce.journal_entries USING btree (reference_id);


--
-- Name: idx_journal_entries_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_journal_entries_updated_at ON thinkce.journal_entries USING btree (updated_at);


--
-- Name: idx_journal_entries_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_journal_entries_version ON thinkce.journal_entries USING btree (version);


--
-- Name: idx_kds_tickets_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_kds_tickets_branch_id ON thinkce.kds_tickets USING btree (branch_id);


--
-- Name: idx_kds_tickets_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_kds_tickets_created_at ON thinkce.kds_tickets USING btree (created_at);


--
-- Name: idx_kds_tickets_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_kds_tickets_deleted_at ON thinkce.kds_tickets USING btree (deleted_at);


--
-- Name: idx_kds_tickets_order_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_kds_tickets_order_id ON thinkce.kds_tickets USING btree (order_id);


--
-- Name: idx_kds_tickets_status; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_kds_tickets_status ON thinkce.kds_tickets USING btree (status);


--
-- Name: idx_kds_tickets_table_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_kds_tickets_table_id ON thinkce.kds_tickets USING btree (table_id);


--
-- Name: idx_kds_tickets_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_kds_tickets_updated_at ON thinkce.kds_tickets USING btree (updated_at);


--
-- Name: idx_kds_tickets_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_kds_tickets_version ON thinkce.kds_tickets USING btree (version);


--
-- Name: idx_leave_requests_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_leave_requests_created_at ON thinkce.leave_requests USING btree (created_at);


--
-- Name: idx_leave_requests_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_leave_requests_deleted_at ON thinkce.leave_requests USING btree (deleted_at);


--
-- Name: idx_leave_requests_staff_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_leave_requests_staff_id ON thinkce.leave_requests USING btree (staff_id);


--
-- Name: idx_leave_requests_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_leave_requests_updated_at ON thinkce.leave_requests USING btree (updated_at);


--
-- Name: idx_leave_requests_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_leave_requests_version ON thinkce.leave_requests USING btree (version);


--
-- Name: idx_ledger_accounts_code; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_ledger_accounts_code ON thinkce.ledger_accounts USING btree (code);


--
-- Name: idx_ledger_accounts_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_accounts_created_at ON thinkce.ledger_accounts USING btree (created_at);


--
-- Name: idx_ledger_accounts_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_accounts_deleted_at ON thinkce.ledger_accounts USING btree (deleted_at);


--
-- Name: idx_ledger_accounts_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_accounts_updated_at ON thinkce.ledger_accounts USING btree (updated_at);


--
-- Name: idx_ledger_accounts_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_accounts_version ON thinkce.ledger_accounts USING btree (version);


--
-- Name: idx_ledger_lines_account_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_lines_account_id ON thinkce.ledger_lines USING btree (account_id);


--
-- Name: idx_ledger_lines_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_lines_created_at ON thinkce.ledger_lines USING btree (created_at);


--
-- Name: idx_ledger_lines_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_lines_deleted_at ON thinkce.ledger_lines USING btree (deleted_at);


--
-- Name: idx_ledger_lines_journal_entry_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_lines_journal_entry_id ON thinkce.ledger_lines USING btree (journal_entry_id);


--
-- Name: idx_ledger_lines_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_lines_updated_at ON thinkce.ledger_lines USING btree (updated_at);


--
-- Name: idx_ledger_lines_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_ledger_lines_version ON thinkce.ledger_lines USING btree (version);


--
-- Name: idx_loyalty_transactions_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_loyalty_transactions_created_at ON thinkce.loyalty_transactions USING btree (created_at);


--
-- Name: idx_loyalty_transactions_customer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_loyalty_transactions_customer_id ON thinkce.loyalty_transactions USING btree (customer_id);


--
-- Name: idx_loyalty_transactions_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_loyalty_transactions_deleted_at ON thinkce.loyalty_transactions USING btree (deleted_at);


--
-- Name: idx_loyalty_transactions_tenant_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_loyalty_transactions_tenant_id ON thinkce.loyalty_transactions USING btree (tenant_id);


--
-- Name: idx_loyalty_transactions_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_loyalty_transactions_updated_at ON thinkce.loyalty_transactions USING btree (updated_at);


--
-- Name: idx_loyalty_transactions_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_loyalty_transactions_version ON thinkce.loyalty_transactions USING btree (version);


--
-- Name: idx_marketing_campaigns_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_marketing_campaigns_created_at ON thinkce.marketing_campaigns USING btree (created_at);


--
-- Name: idx_marketing_campaigns_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_marketing_campaigns_deleted_at ON thinkce.marketing_campaigns USING btree (deleted_at);


--
-- Name: idx_marketing_campaigns_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_marketing_campaigns_updated_at ON thinkce.marketing_campaigns USING btree (updated_at);


--
-- Name: idx_marketing_campaigns_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_marketing_campaigns_version ON thinkce.marketing_campaigns USING btree (version);


--
-- Name: idx_newsletter_subscriptions_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_newsletter_subscriptions_created_at ON thinkce.newsletter_subscriptions USING btree (created_at);


--
-- Name: idx_newsletter_subscriptions_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_newsletter_subscriptions_deleted_at ON thinkce.newsletter_subscriptions USING btree (deleted_at);


--
-- Name: idx_newsletter_subscriptions_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_newsletter_subscriptions_updated_at ON thinkce.newsletter_subscriptions USING btree (updated_at);


--
-- Name: idx_newsletter_subscriptions_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_newsletter_subscriptions_version ON thinkce.newsletter_subscriptions USING btree (version);


--
-- Name: idx_notification_settings_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notification_settings_created_at ON thinkce.notification_settings USING btree (created_at);


--
-- Name: idx_notification_settings_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notification_settings_deleted_at ON thinkce.notification_settings USING btree (deleted_at);


--
-- Name: idx_notification_settings_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notification_settings_updated_at ON thinkce.notification_settings USING btree (updated_at);


--
-- Name: idx_notification_settings_user_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_notification_settings_user_id ON thinkce.notification_settings USING btree (user_id);


--
-- Name: idx_notification_settings_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notification_settings_version ON thinkce.notification_settings USING btree (version);


--
-- Name: idx_notifications_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notifications_created_at ON thinkce.notifications USING btree (created_at);


--
-- Name: idx_notifications_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notifications_deleted_at ON thinkce.notifications USING btree (deleted_at);


--
-- Name: idx_notifications_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notifications_updated_at ON thinkce.notifications USING btree (updated_at);


--
-- Name: idx_notifications_user_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notifications_user_id ON thinkce.notifications USING btree (user_id);


--
-- Name: idx_notifications_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_notifications_version ON thinkce.notifications USING btree (version);


--
-- Name: idx_order_items_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_order_items_created_at ON thinkce.order_items USING btree (created_at);


--
-- Name: idx_order_items_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_order_items_deleted_at ON thinkce.order_items USING btree (deleted_at);


--
-- Name: idx_order_items_order_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_order_items_order_id ON thinkce.order_items USING btree (order_id);


--
-- Name: idx_order_items_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_order_items_product_id ON thinkce.order_items USING btree (product_id);


--
-- Name: idx_order_items_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_order_items_updated_at ON thinkce.order_items USING btree (updated_at);


--
-- Name: idx_order_items_variant_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_order_items_variant_id ON thinkce.order_items USING btree (variant_id);


--
-- Name: idx_order_items_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_order_items_version ON thinkce.order_items USING btree (version);


--
-- Name: idx_orders_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_orders_branch_id ON thinkce.orders USING btree (branch_id);


--
-- Name: idx_orders_cashier_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_orders_cashier_id ON thinkce.orders USING btree (cashier_id);


--
-- Name: idx_orders_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_orders_created_at ON thinkce.orders USING btree (created_at);


--
-- Name: idx_orders_customer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_orders_customer_id ON thinkce.orders USING btree (customer_id);


--
-- Name: idx_orders_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_orders_deleted_at ON thinkce.orders USING btree (deleted_at);


--
-- Name: idx_orders_order_number; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_orders_order_number ON thinkce.orders USING btree (order_number);


--
-- Name: idx_orders_receipt_token; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_orders_receipt_token ON thinkce.orders USING btree (receipt_token);


--
-- Name: idx_orders_status; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_orders_status ON thinkce.orders USING btree (status);


--
-- Name: idx_orders_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_orders_updated_at ON thinkce.orders USING btree (updated_at);


--
-- Name: idx_orders_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_orders_version ON thinkce.orders USING btree (version);


--
-- Name: idx_payment_methods_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payment_methods_created_at ON thinkce.payment_methods USING btree (created_at);


--
-- Name: idx_payment_methods_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payment_methods_deleted_at ON thinkce.payment_methods USING btree (deleted_at);


--
-- Name: idx_payment_methods_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payment_methods_updated_at ON thinkce.payment_methods USING btree (updated_at);


--
-- Name: idx_payment_methods_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payment_methods_version ON thinkce.payment_methods USING btree (version);


--
-- Name: idx_payments_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payments_created_at ON thinkce.payments USING btree (created_at);


--
-- Name: idx_payments_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payments_deleted_at ON thinkce.payments USING btree (deleted_at);


--
-- Name: idx_payments_order_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payments_order_id ON thinkce.payments USING btree (order_id);


--
-- Name: idx_payments_payment_method_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payments_payment_method_id ON thinkce.payments USING btree (payment_method_id);


--
-- Name: idx_payments_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payments_updated_at ON thinkce.payments USING btree (updated_at);


--
-- Name: idx_payments_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payments_version ON thinkce.payments USING btree (version);


--
-- Name: idx_payroll_periods_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_periods_created_at ON thinkce.payroll_periods USING btree (created_at);


--
-- Name: idx_payroll_periods_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_periods_deleted_at ON thinkce.payroll_periods USING btree (deleted_at);


--
-- Name: idx_payroll_periods_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_periods_updated_at ON thinkce.payroll_periods USING btree (updated_at);


--
-- Name: idx_payroll_periods_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_periods_version ON thinkce.payroll_periods USING btree (version);


--
-- Name: idx_payroll_records_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_records_created_at ON thinkce.payroll_records USING btree (created_at);


--
-- Name: idx_payroll_records_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_records_deleted_at ON thinkce.payroll_records USING btree (deleted_at);


--
-- Name: idx_payroll_records_is_paid; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_records_is_paid ON thinkce.payroll_records USING btree (is_paid);


--
-- Name: idx_payroll_records_period_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_records_period_id ON thinkce.payroll_records USING btree (period_id);


--
-- Name: idx_payroll_records_staff_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_records_staff_id ON thinkce.payroll_records USING btree (staff_id);


--
-- Name: idx_payroll_records_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_records_updated_at ON thinkce.payroll_records USING btree (updated_at);


--
-- Name: idx_payroll_records_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_payroll_records_version ON thinkce.payroll_records USING btree (version);


--
-- Name: idx_print_jobs_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_print_jobs_branch_id ON thinkce.print_jobs USING btree (branch_id);


--
-- Name: idx_print_jobs_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_print_jobs_created_at ON thinkce.print_jobs USING btree (created_at);


--
-- Name: idx_print_jobs_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_print_jobs_deleted_at ON thinkce.print_jobs USING btree (deleted_at);


--
-- Name: idx_print_jobs_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_print_jobs_updated_at ON thinkce.print_jobs USING btree (updated_at);


--
-- Name: idx_print_jobs_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_print_jobs_version ON thinkce.print_jobs USING btree (version);


--
-- Name: idx_product_components_component_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_components_component_product_id ON thinkce.product_components USING btree (component_product_id);


--
-- Name: idx_product_components_composite_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_components_composite_product_id ON thinkce.product_components USING btree (composite_product_id);


--
-- Name: idx_product_components_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_components_created_at ON thinkce.product_components USING btree (created_at);


--
-- Name: idx_product_components_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_components_deleted_at ON thinkce.product_components USING btree (deleted_at);


--
-- Name: idx_product_components_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_components_updated_at ON thinkce.product_components USING btree (updated_at);


--
-- Name: idx_product_components_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_components_version ON thinkce.product_components USING btree (version);


--
-- Name: idx_product_histories_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_histories_created_at ON thinkce.product_histories USING btree (created_at);


--
-- Name: idx_product_histories_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_histories_deleted_at ON thinkce.product_histories USING btree (deleted_at);


--
-- Name: idx_product_histories_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_histories_product_id ON thinkce.product_histories USING btree (product_id);


--
-- Name: idx_product_histories_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_histories_updated_at ON thinkce.product_histories USING btree (updated_at);


--
-- Name: idx_product_histories_user_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_histories_user_id ON thinkce.product_histories USING btree (user_id);


--
-- Name: idx_product_histories_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_histories_version ON thinkce.product_histories USING btree (version);


--
-- Name: idx_product_image_galleries_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_image_galleries_created_at ON thinkce.product_image_galleries USING btree (created_at);


--
-- Name: idx_product_image_galleries_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_image_galleries_deleted_at ON thinkce.product_image_galleries USING btree (deleted_at);


--
-- Name: idx_product_image_galleries_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_image_galleries_product_id ON thinkce.product_image_galleries USING btree (product_id);


--
-- Name: idx_product_image_galleries_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_image_galleries_updated_at ON thinkce.product_image_galleries USING btree (updated_at);


--
-- Name: idx_product_image_galleries_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_image_galleries_version ON thinkce.product_image_galleries USING btree (version);


--
-- Name: idx_product_reviews_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_reviews_created_at ON thinkce.product_reviews USING btree (created_at);


--
-- Name: idx_product_reviews_customer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_reviews_customer_id ON thinkce.product_reviews USING btree (customer_id);


--
-- Name: idx_product_reviews_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_reviews_deleted_at ON thinkce.product_reviews USING btree (deleted_at);


--
-- Name: idx_product_reviews_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_reviews_product_id ON thinkce.product_reviews USING btree (product_id);


--
-- Name: idx_product_reviews_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_reviews_updated_at ON thinkce.product_reviews USING btree (updated_at);


--
-- Name: idx_product_reviews_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_reviews_version ON thinkce.product_reviews USING btree (version);


--
-- Name: idx_product_variants_barcode; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_variants_barcode ON thinkce.product_variants USING btree (barcode);


--
-- Name: idx_product_variants_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_variants_created_at ON thinkce.product_variants USING btree (created_at);


--
-- Name: idx_product_variants_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_variants_deleted_at ON thinkce.product_variants USING btree (deleted_at);


--
-- Name: idx_product_variants_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_variants_product_id ON thinkce.product_variants USING btree (product_id);


--
-- Name: idx_product_variants_sku; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_product_variants_sku ON thinkce.product_variants USING btree (sku);


--
-- Name: idx_product_variants_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_variants_updated_at ON thinkce.product_variants USING btree (updated_at);


--
-- Name: idx_product_variants_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_product_variants_version ON thinkce.product_variants USING btree (version);


--
-- Name: idx_products_barcode; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_barcode ON thinkce.products USING btree (barcode);


--
-- Name: idx_products_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_branch_id ON thinkce.products USING btree (branch_id);


--
-- Name: idx_products_category_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_category_id ON thinkce.products USING btree (category_id);


--
-- Name: idx_products_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_created_at ON thinkce.products USING btree (created_at);


--
-- Name: idx_products_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_deleted_at ON thinkce.products USING btree (deleted_at);


--
-- Name: idx_products_has_variants; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_has_variants ON thinkce.products USING btree (has_variants);


--
-- Name: idx_products_is_active; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_is_active ON thinkce.products USING btree (is_active);


--
-- Name: idx_products_is_online; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_is_online ON thinkce.products USING btree (is_online);


--
-- Name: idx_products_sku; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_products_sku ON thinkce.products USING btree (sku);


--
-- Name: idx_products_supplier_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_supplier_id ON thinkce.products USING btree (supplier_id);


--
-- Name: idx_products_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_updated_at ON thinkce.products USING btree (updated_at);


--
-- Name: idx_products_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_products_version ON thinkce.products USING btree (version);


--
-- Name: idx_promotions_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_promotions_deleted_at ON thinkce.promotions USING btree (deleted_at);


--
-- Name: idx_purchase_order_items_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_order_items_created_at ON thinkce.purchase_order_items USING btree (created_at);


--
-- Name: idx_purchase_order_items_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_order_items_deleted_at ON thinkce.purchase_order_items USING btree (deleted_at);


--
-- Name: idx_purchase_order_items_po_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_order_items_po_id ON thinkce.purchase_order_items USING btree (po_id);


--
-- Name: idx_purchase_order_items_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_order_items_product_id ON thinkce.purchase_order_items USING btree (product_id);


--
-- Name: idx_purchase_order_items_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_order_items_updated_at ON thinkce.purchase_order_items USING btree (updated_at);


--
-- Name: idx_purchase_order_items_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_order_items_version ON thinkce.purchase_order_items USING btree (version);


--
-- Name: idx_purchase_orders_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_orders_branch_id ON thinkce.purchase_orders USING btree (branch_id);


--
-- Name: idx_purchase_orders_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_orders_created_at ON thinkce.purchase_orders USING btree (created_at);


--
-- Name: idx_purchase_orders_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_orders_deleted_at ON thinkce.purchase_orders USING btree (deleted_at);


--
-- Name: idx_purchase_orders_po_number; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_purchase_orders_po_number ON thinkce.purchase_orders USING btree (po_number);


--
-- Name: idx_purchase_orders_status; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_orders_status ON thinkce.purchase_orders USING btree (status);


--
-- Name: idx_purchase_orders_supplier_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_orders_supplier_id ON thinkce.purchase_orders USING btree (supplier_id);


--
-- Name: idx_purchase_orders_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_orders_updated_at ON thinkce.purchase_orders USING btree (updated_at);


--
-- Name: idx_purchase_orders_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_purchase_orders_version ON thinkce.purchase_orders USING btree (version);


--
-- Name: idx_quotation_items_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotation_items_created_at ON thinkce.quotation_items USING btree (created_at);


--
-- Name: idx_quotation_items_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotation_items_deleted_at ON thinkce.quotation_items USING btree (deleted_at);


--
-- Name: idx_quotation_items_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotation_items_product_id ON thinkce.quotation_items USING btree (product_id);


--
-- Name: idx_quotation_items_quotation_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotation_items_quotation_id ON thinkce.quotation_items USING btree (quotation_id);


--
-- Name: idx_quotation_items_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotation_items_updated_at ON thinkce.quotation_items USING btree (updated_at);


--
-- Name: idx_quotation_items_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotation_items_version ON thinkce.quotation_items USING btree (version);


--
-- Name: idx_quotations_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotations_branch_id ON thinkce.quotations USING btree (branch_id);


--
-- Name: idx_quotations_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotations_created_at ON thinkce.quotations USING btree (created_at);


--
-- Name: idx_quotations_customer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotations_customer_id ON thinkce.quotations USING btree (customer_id);


--
-- Name: idx_quotations_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotations_deleted_at ON thinkce.quotations USING btree (deleted_at);


--
-- Name: idx_quotations_quote_number; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_quotations_quote_number ON thinkce.quotations USING btree (quote_number);


--
-- Name: idx_quotations_status; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotations_status ON thinkce.quotations USING btree (status);


--
-- Name: idx_quotations_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotations_updated_at ON thinkce.quotations USING btree (updated_at);


--
-- Name: idx_quotations_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_quotations_version ON thinkce.quotations USING btree (version);


--
-- Name: idx_return_items_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_return_items_created_at ON thinkce.return_items USING btree (created_at);


--
-- Name: idx_return_items_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_return_items_deleted_at ON thinkce.return_items USING btree (deleted_at);


--
-- Name: idx_return_items_return_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_return_items_return_id ON thinkce.return_items USING btree (return_id);


--
-- Name: idx_return_items_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_return_items_updated_at ON thinkce.return_items USING btree (updated_at);


--
-- Name: idx_return_items_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_return_items_version ON thinkce.return_items USING btree (version);


--
-- Name: idx_returns_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_returns_branch_id ON thinkce.returns USING btree (branch_id);


--
-- Name: idx_returns_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_returns_created_at ON thinkce.returns USING btree (created_at);


--
-- Name: idx_returns_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_returns_deleted_at ON thinkce.returns USING btree (deleted_at);


--
-- Name: idx_returns_order_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_returns_order_id ON thinkce.returns USING btree (order_id);


--
-- Name: idx_returns_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_returns_updated_at ON thinkce.returns USING btree (updated_at);


--
-- Name: idx_returns_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_returns_version ON thinkce.returns USING btree (version);


--
-- Name: idx_service_categories_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_categories_created_at ON thinkce.service_categories USING btree (created_at);


--
-- Name: idx_service_categories_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_categories_deleted_at ON thinkce.service_categories USING btree (deleted_at);


--
-- Name: idx_service_categories_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_categories_updated_at ON thinkce.service_categories USING btree (updated_at);


--
-- Name: idx_service_categories_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_categories_version ON thinkce.service_categories USING btree (version);


--
-- Name: idx_service_commission_records_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_records_created_at ON thinkce.service_commission_records USING btree (created_at);


--
-- Name: idx_service_commission_records_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_records_deleted_at ON thinkce.service_commission_records USING btree (deleted_at);


--
-- Name: idx_service_commission_records_is_paid; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_records_is_paid ON thinkce.service_commission_records USING btree (is_paid);


--
-- Name: idx_service_commission_records_order_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_records_order_id ON thinkce.service_commission_records USING btree (order_id);


--
-- Name: idx_service_commission_records_staff_member_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_records_staff_member_id ON thinkce.service_commission_records USING btree (staff_member_id);


--
-- Name: idx_service_commission_records_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_records_updated_at ON thinkce.service_commission_records USING btree (updated_at);


--
-- Name: idx_service_commission_records_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_records_version ON thinkce.service_commission_records USING btree (version);


--
-- Name: idx_service_commission_rules_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_rules_created_at ON thinkce.service_commission_rules USING btree (created_at);


--
-- Name: idx_service_commission_rules_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_rules_deleted_at ON thinkce.service_commission_rules USING btree (deleted_at);


--
-- Name: idx_service_commission_rules_staff_member_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_rules_staff_member_id ON thinkce.service_commission_rules USING btree (staff_member_id);


--
-- Name: idx_service_commission_rules_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_rules_updated_at ON thinkce.service_commission_rules USING btree (updated_at);


--
-- Name: idx_service_commission_rules_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_service_commission_rules_version ON thinkce.service_commission_rules USING btree (version);


--
-- Name: idx_services_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_services_branch_id ON thinkce.services USING btree (branch_id);


--
-- Name: idx_services_category_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_services_category_id ON thinkce.services USING btree (category_id);


--
-- Name: idx_services_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_services_created_at ON thinkce.services USING btree (created_at);


--
-- Name: idx_services_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_services_deleted_at ON thinkce.services USING btree (deleted_at);


--
-- Name: idx_services_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_services_updated_at ON thinkce.services USING btree (updated_at);


--
-- Name: idx_services_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_services_version ON thinkce.services USING btree (version);


--
-- Name: idx_shifts_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_shifts_branch_id ON thinkce.shifts USING btree (branch_id);


--
-- Name: idx_shifts_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_shifts_created_at ON thinkce.shifts USING btree (created_at);


--
-- Name: idx_shifts_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_shifts_deleted_at ON thinkce.shifts USING btree (deleted_at);


--
-- Name: idx_shifts_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_shifts_updated_at ON thinkce.shifts USING btree (updated_at);


--
-- Name: idx_shifts_user_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_shifts_user_id ON thinkce.shifts USING btree (user_id);


--
-- Name: idx_shifts_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_shifts_version ON thinkce.shifts USING btree (version);


--
-- Name: idx_split_bill_groups_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_split_bill_groups_branch_id ON thinkce.split_bill_groups USING btree (branch_id);


--
-- Name: idx_split_bill_groups_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_split_bill_groups_created_at ON thinkce.split_bill_groups USING btree (created_at);


--
-- Name: idx_split_bill_groups_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_split_bill_groups_deleted_at ON thinkce.split_bill_groups USING btree (deleted_at);


--
-- Name: idx_split_bill_groups_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_split_bill_groups_updated_at ON thinkce.split_bill_groups USING btree (updated_at);


--
-- Name: idx_split_bill_groups_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_split_bill_groups_version ON thinkce.split_bill_groups USING btree (version);


--
-- Name: idx_stock_batches_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_batches_branch_id ON thinkce.stock_batches USING btree (branch_id);


--
-- Name: idx_stock_batches_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_batches_created_at ON thinkce.stock_batches USING btree (created_at);


--
-- Name: idx_stock_batches_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_batches_deleted_at ON thinkce.stock_batches USING btree (deleted_at);


--
-- Name: idx_stock_batches_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_batches_product_id ON thinkce.stock_batches USING btree (product_id);


--
-- Name: idx_stock_batches_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_batches_updated_at ON thinkce.stock_batches USING btree (updated_at);


--
-- Name: idx_stock_batches_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_batches_version ON thinkce.stock_batches USING btree (version);


--
-- Name: idx_stock_movements_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_movements_branch_id ON thinkce.stock_movements USING btree (branch_id);


--
-- Name: idx_stock_movements_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_movements_created_at ON thinkce.stock_movements USING btree (created_at);


--
-- Name: idx_stock_movements_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_movements_deleted_at ON thinkce.stock_movements USING btree (deleted_at);


--
-- Name: idx_stock_movements_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_movements_product_id ON thinkce.stock_movements USING btree (product_id);


--
-- Name: idx_stock_movements_tenant_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_movements_tenant_id ON thinkce.stock_movements USING btree (tenant_id);


--
-- Name: idx_stock_movements_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_movements_updated_at ON thinkce.stock_movements USING btree (updated_at);


--
-- Name: idx_stock_movements_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_movements_version ON thinkce.stock_movements USING btree (version);


--
-- Name: idx_stock_transfer_items_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfer_items_created_at ON thinkce.stock_transfer_items USING btree (created_at);


--
-- Name: idx_stock_transfer_items_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfer_items_deleted_at ON thinkce.stock_transfer_items USING btree (deleted_at);


--
-- Name: idx_stock_transfer_items_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfer_items_product_id ON thinkce.stock_transfer_items USING btree (product_id);


--
-- Name: idx_stock_transfer_items_transfer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfer_items_transfer_id ON thinkce.stock_transfer_items USING btree (transfer_id);


--
-- Name: idx_stock_transfer_items_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfer_items_updated_at ON thinkce.stock_transfer_items USING btree (updated_at);


--
-- Name: idx_stock_transfer_items_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfer_items_version ON thinkce.stock_transfer_items USING btree (version);


--
-- Name: idx_stock_transfers_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfers_created_at ON thinkce.stock_transfers USING btree (created_at);


--
-- Name: idx_stock_transfers_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfers_deleted_at ON thinkce.stock_transfers USING btree (deleted_at);


--
-- Name: idx_stock_transfers_from_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfers_from_branch_id ON thinkce.stock_transfers USING btree (from_branch_id);


--
-- Name: idx_stock_transfers_reference_no; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_stock_transfers_reference_no ON thinkce.stock_transfers USING btree (reference_no);


--
-- Name: idx_stock_transfers_status; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfers_status ON thinkce.stock_transfers USING btree (status);


--
-- Name: idx_stock_transfers_to_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfers_to_branch_id ON thinkce.stock_transfers USING btree (to_branch_id);


--
-- Name: idx_stock_transfers_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfers_updated_at ON thinkce.stock_transfers USING btree (updated_at);


--
-- Name: idx_stock_transfers_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stock_transfers_version ON thinkce.stock_transfers USING btree (version);


--
-- Name: idx_stocktake_entries_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_entries_created_at ON thinkce.stocktake_entries USING btree (created_at);


--
-- Name: idx_stocktake_entries_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_entries_deleted_at ON thinkce.stocktake_entries USING btree (deleted_at);


--
-- Name: idx_stocktake_entries_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_entries_product_id ON thinkce.stocktake_entries USING btree (product_id);


--
-- Name: idx_stocktake_entries_session_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_entries_session_id ON thinkce.stocktake_entries USING btree (session_id);


--
-- Name: idx_stocktake_entries_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_entries_updated_at ON thinkce.stocktake_entries USING btree (updated_at);


--
-- Name: idx_stocktake_entries_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_entries_version ON thinkce.stocktake_entries USING btree (version);


--
-- Name: idx_stocktake_sessions_access_token; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_stocktake_sessions_access_token ON thinkce.stocktake_sessions USING btree (access_token);


--
-- Name: idx_stocktake_sessions_branch_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_sessions_branch_id ON thinkce.stocktake_sessions USING btree (branch_id);


--
-- Name: idx_stocktake_sessions_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_sessions_created_at ON thinkce.stocktake_sessions USING btree (created_at);


--
-- Name: idx_stocktake_sessions_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_sessions_deleted_at ON thinkce.stocktake_sessions USING btree (deleted_at);


--
-- Name: idx_stocktake_sessions_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_sessions_updated_at ON thinkce.stocktake_sessions USING btree (updated_at);


--
-- Name: idx_stocktake_sessions_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_stocktake_sessions_version ON thinkce.stocktake_sessions USING btree (version);


--
-- Name: idx_storefront_settings_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_storefront_settings_created_at ON thinkce.storefront_settings USING btree (created_at);


--
-- Name: idx_storefront_settings_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_storefront_settings_deleted_at ON thinkce.storefront_settings USING btree (deleted_at);


--
-- Name: idx_storefront_settings_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_storefront_settings_updated_at ON thinkce.storefront_settings USING btree (updated_at);


--
-- Name: idx_storefront_settings_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_storefront_settings_version ON thinkce.storefront_settings USING btree (version);


--
-- Name: idx_supplier_ledger_entries_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_created_at ON thinkce.supplier_ledger_entries USING btree (created_at);


--
-- Name: idx_supplier_ledger_entries_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_deleted_at ON thinkce.supplier_ledger_entries USING btree (deleted_at);


--
-- Name: idx_supplier_ledger_entries_supplier_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_supplier_id ON thinkce.supplier_ledger_entries USING btree (supplier_id);


--
-- Name: idx_supplier_ledger_entries_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_updated_at ON thinkce.supplier_ledger_entries USING btree (updated_at);


--
-- Name: idx_supplier_ledger_entries_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_ledger_entries_version ON thinkce.supplier_ledger_entries USING btree (version);


--
-- Name: idx_supplier_product; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_supplier_product ON thinkce.supplier_products USING btree (supplier_id, product_id);


--
-- Name: idx_supplier_products_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_products_created_at ON thinkce.supplier_products USING btree (created_at);


--
-- Name: idx_supplier_products_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_products_deleted_at ON thinkce.supplier_products USING btree (deleted_at);


--
-- Name: idx_supplier_products_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_products_product_id ON thinkce.supplier_products USING btree (product_id);


--
-- Name: idx_supplier_products_supplier_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_products_supplier_id ON thinkce.supplier_products USING btree (supplier_id);


--
-- Name: idx_supplier_products_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_products_updated_at ON thinkce.supplier_products USING btree (updated_at);


--
-- Name: idx_supplier_products_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_products_version ON thinkce.supplier_products USING btree (version);


--
-- Name: idx_supplier_profiles_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_profiles_created_at ON thinkce.supplier_profiles USING btree (created_at);


--
-- Name: idx_supplier_profiles_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_profiles_deleted_at ON thinkce.supplier_profiles USING btree (deleted_at);


--
-- Name: idx_supplier_profiles_supplier_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_profiles_supplier_id ON thinkce.supplier_profiles USING btree (supplier_id);


--
-- Name: idx_supplier_profiles_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_profiles_updated_at ON thinkce.supplier_profiles USING btree (updated_at);


--
-- Name: idx_supplier_profiles_user_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_profiles_user_id ON thinkce.supplier_profiles USING btree (user_id);


--
-- Name: idx_supplier_profiles_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_supplier_profiles_version ON thinkce.supplier_profiles USING btree (version);


--
-- Name: idx_suppliers_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_suppliers_created_at ON thinkce.suppliers USING btree (created_at);


--
-- Name: idx_suppliers_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_suppliers_deleted_at ON thinkce.suppliers USING btree (deleted_at);


--
-- Name: idx_suppliers_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_suppliers_updated_at ON thinkce.suppliers USING btree (updated_at);


--
-- Name: idx_suppliers_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_suppliers_version ON thinkce.suppliers USING btree (version);


--
-- Name: idx_tax_configurations_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_tax_configurations_created_at ON thinkce.tax_configurations USING btree (created_at);


--
-- Name: idx_tax_configurations_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_tax_configurations_deleted_at ON thinkce.tax_configurations USING btree (deleted_at);


--
-- Name: idx_tax_configurations_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_tax_configurations_updated_at ON thinkce.tax_configurations USING btree (updated_at);


--
-- Name: idx_tax_configurations_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_tax_configurations_version ON thinkce.tax_configurations USING btree (version);


--
-- Name: idx_tenant_branch_sku; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_branch_sku ON thinkce.products USING btree (branch_id, sku);


--
-- Name: idx_tenant_integrations_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_tenant_integrations_created_at ON thinkce.tenant_integrations USING btree (created_at);


--
-- Name: idx_tenant_integrations_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_tenant_integrations_deleted_at ON thinkce.tenant_integrations USING btree (deleted_at);


--
-- Name: idx_tenant_integrations_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_tenant_integrations_updated_at ON thinkce.tenant_integrations USING btree (updated_at);


--
-- Name: idx_tenant_integrations_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_tenant_integrations_version ON thinkce.tenant_integrations USING btree (version);


--
-- Name: idx_tenant_provider; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_provider ON thinkce.tenant_integrations USING btree (provider);


--
-- Name: idx_webhook_endpoints_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_endpoints_created_at ON thinkce.webhook_endpoints USING btree (created_at);


--
-- Name: idx_webhook_endpoints_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_endpoints_deleted_at ON thinkce.webhook_endpoints USING btree (deleted_at);


--
-- Name: idx_webhook_endpoints_external_system_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_endpoints_external_system_id ON thinkce.webhook_endpoints USING btree (external_system_id);


--
-- Name: idx_webhook_endpoints_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_endpoints_updated_at ON thinkce.webhook_endpoints USING btree (updated_at);


--
-- Name: idx_webhook_endpoints_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_endpoints_version ON thinkce.webhook_endpoints USING btree (version);


--
-- Name: idx_webhook_events_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_events_created_at ON thinkce.webhook_events USING btree (created_at);


--
-- Name: idx_webhook_events_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_events_deleted_at ON thinkce.webhook_events USING btree (deleted_at);


--
-- Name: idx_webhook_events_endpoint_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_events_endpoint_id ON thinkce.webhook_events USING btree (endpoint_id);


--
-- Name: idx_webhook_events_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_events_updated_at ON thinkce.webhook_events USING btree (updated_at);


--
-- Name: idx_webhook_events_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_webhook_events_version ON thinkce.webhook_events USING btree (version);


--
-- Name: idx_wishlists_created_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_wishlists_created_at ON thinkce.wishlists USING btree (created_at);


--
-- Name: idx_wishlists_customer_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_wishlists_customer_id ON thinkce.wishlists USING btree (customer_id);


--
-- Name: idx_wishlists_deleted_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_wishlists_deleted_at ON thinkce.wishlists USING btree (deleted_at);


--
-- Name: idx_wishlists_product_id; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_wishlists_product_id ON thinkce.wishlists USING btree (product_id);


--
-- Name: idx_wishlists_updated_at; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_wishlists_updated_at ON thinkce.wishlists USING btree (updated_at);


--
-- Name: idx_wishlists_version; Type: INDEX; Schema: thinkce; Owner: -
--

CREATE INDEX idx_wishlists_version ON thinkce.wishlists USING btree (version);


--
-- Name: admin_users fk_admin_users_role; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT fk_admin_users_role FOREIGN KEY (admin_role_id) REFERENCES public.admin_roles(id);


--
-- Name: admin_users fk_admin_users_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_users
    ADD CONSTRAINT fk_admin_users_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: audit_logs fk_audit_logs_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: audit_logs fk_audit_logs_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: billing_payments fk_billing_payments_subscription; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_payments
    ADD CONSTRAINT fk_billing_payments_subscription FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id);


--
-- Name: blog_posts fk_blog_posts_author; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blog_posts
    ADD CONSTRAINT fk_blog_posts_author FOREIGN KEY (author_id) REFERENCES public.users(id);


--
-- Name: branches fk_branches_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.branches
    ADD CONSTRAINT fk_branches_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: plan_features fk_pricing_plans_features; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features
    ADD CONSTRAINT fk_pricing_plans_features FOREIGN KEY (plan_id) REFERENCES public.pricing_plans(id);


--
-- Name: domains fk_public_tenants_domains; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.domains
    ADD CONSTRAINT fk_public_tenants_domains FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: tenants fk_public_tenants_referred_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT fk_public_tenants_referred_by FOREIGN KEY (referred_by_id) REFERENCES public.tenants(id);


--
-- Name: subscriptions fk_public_tenants_subscription; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT fk_public_tenants_subscription FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: user_profiles fk_public_user_profiles_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT fk_public_user_profiles_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: user_profiles fk_public_users_profiles; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT fk_public_users_profiles FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: referral_rewards fk_referral_rewards_referred_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_rewards
    ADD CONSTRAINT fk_referral_rewards_referred_tenant FOREIGN KEY (referred_tenant_id) REFERENCES public.tenants(id);


--
-- Name: referral_rewards fk_referral_rewards_referrer; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.referral_rewards
    ADD CONSTRAINT fk_referral_rewards_referrer FOREIGN KEY (referrer_id) REFERENCES public.tenants(id);


--
-- Name: seo_settings fk_seo_settings_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seo_settings
    ADD CONSTRAINT fk_seo_settings_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: tenant_metrics fk_tenant_metrics_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_metrics
    ADD CONSTRAINT fk_tenant_metrics_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: api_keys fk_api_keys_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.api_keys
    ADD CONSTRAINT fk_api_keys_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: api_request_logs fk_api_request_logs_tenant; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.api_request_logs
    ADD CONSTRAINT fk_api_request_logs_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: appointments fk_appointments_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.appointments
    ADD CONSTRAINT fk_appointments_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: attendances fk_attendances_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.attendances
    ADD CONSTRAINT fk_attendances_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: audit_logs fk_audit_logs_tenant; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.audit_logs
    ADD CONSTRAINT fk_audit_logs_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: audit_logs fk_audit_logs_user; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.audit_logs
    ADD CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: branches fk_branches_tenant; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.branches
    ADD CONSTRAINT fk_branches_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: cash_drawer_sessions fk_cash_drawer_sessions_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.cash_drawer_sessions
    ADD CONSTRAINT fk_cash_drawer_sessions_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: commission_rules fk_commission_rules_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.commission_rules
    ADD CONSTRAINT fk_commission_rules_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: customers fk_customers_tier; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.customers
    ADD CONSTRAINT fk_customers_tier FOREIGN KEY (tier_id) REFERENCES softivite.customer_tiers(id);


--
-- Name: delivery_orders fk_delivery_orders_driver; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.delivery_orders
    ADD CONSTRAINT fk_delivery_orders_driver FOREIGN KEY (driver_id) REFERENCES softivite.delivery_drivers(id);


--
-- Name: delivery_orders fk_delivery_orders_order; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.delivery_orders
    ADD CONSTRAINT fk_delivery_orders_order FOREIGN KEY (order_id) REFERENCES softivite.orders(id);


--
-- Name: dining_tables fk_dining_tables_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.dining_tables
    ADD CONSTRAINT fk_dining_tables_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: expenses fk_expenses_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.expenses
    ADD CONSTRAINT fk_expenses_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: expenses fk_expenses_category; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.expenses
    ADD CONSTRAINT fk_expenses_category FOREIGN KEY (category_id) REFERENCES softivite.categories(id);


--
-- Name: expenses fk_expenses_created_by; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.expenses
    ADD CONSTRAINT fk_expenses_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: ledger_lines fk_journal_entries_lines; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.ledger_lines
    ADD CONSTRAINT fk_journal_entries_lines FOREIGN KEY (journal_entry_id) REFERENCES softivite.journal_entries(id);


--
-- Name: kds_tickets fk_kds_tickets_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.kds_tickets
    ADD CONSTRAINT fk_kds_tickets_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: kds_tickets fk_kds_tickets_order; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.kds_tickets
    ADD CONSTRAINT fk_kds_tickets_order FOREIGN KEY (order_id) REFERENCES softivite.orders(id);


--
-- Name: kds_tickets fk_kds_tickets_table; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.kds_tickets
    ADD CONSTRAINT fk_kds_tickets_table FOREIGN KEY (table_id) REFERENCES softivite.dining_tables(id);


--
-- Name: leave_requests fk_leave_requests_reviewed_by; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.leave_requests
    ADD CONSTRAINT fk_leave_requests_reviewed_by FOREIGN KEY (reviewed_by_id) REFERENCES public.users(id);


--
-- Name: ledger_lines fk_ledger_lines_account; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.ledger_lines
    ADD CONSTRAINT fk_ledger_lines_account FOREIGN KEY (account_id) REFERENCES softivite.ledger_accounts(id);


--
-- Name: loyalty_transactions fk_loyalty_transactions_tenant; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.loyalty_transactions
    ADD CONSTRAINT fk_loyalty_transactions_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: notifications fk_notifications_user; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.notifications
    ADD CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: order_items fk_order_items_product; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.order_items
    ADD CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES softivite.products(id);


--
-- Name: orders fk_orders_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.orders
    ADD CONSTRAINT fk_orders_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: orders fk_orders_cashier; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.orders
    ADD CONSTRAINT fk_orders_cashier FOREIGN KEY (cashier_id) REFERENCES public.users(id);


--
-- Name: orders fk_orders_customer; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.orders
    ADD CONSTRAINT fk_orders_customer FOREIGN KEY (customer_id) REFERENCES softivite.customers(id);


--
-- Name: order_items fk_orders_items; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.order_items
    ADD CONSTRAINT fk_orders_items FOREIGN KEY (order_id) REFERENCES softivite.orders(id);


--
-- Name: print_jobs fk_print_jobs_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.print_jobs
    ADD CONSTRAINT fk_print_jobs_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: products fk_products_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.products
    ADD CONSTRAINT fk_products_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: products fk_products_category; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.products
    ADD CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES softivite.categories(id);


--
-- Name: products fk_products_supplier; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.products
    ADD CONSTRAINT fk_products_supplier FOREIGN KEY (supplier_id) REFERENCES softivite.suppliers(id);


--
-- Name: domains fk_public_tenants_domains; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.domains
    ADD CONSTRAINT fk_public_tenants_domains FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: purchase_order_items fk_purchase_order_items_product; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.purchase_order_items
    ADD CONSTRAINT fk_purchase_order_items_product FOREIGN KEY (product_id) REFERENCES softivite.products(id);


--
-- Name: purchase_orders fk_purchase_orders_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.purchase_orders
    ADD CONSTRAINT fk_purchase_orders_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: purchase_order_items fk_purchase_orders_items; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.purchase_order_items
    ADD CONSTRAINT fk_purchase_orders_items FOREIGN KEY (po_id) REFERENCES softivite.purchase_orders(id);


--
-- Name: purchase_orders fk_purchase_orders_supplier; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.purchase_orders
    ADD CONSTRAINT fk_purchase_orders_supplier FOREIGN KEY (supplier_id) REFERENCES softivite.suppliers(id);


--
-- Name: quotations fk_quotations_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.quotations
    ADD CONSTRAINT fk_quotations_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: quotations fk_quotations_created_by; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.quotations
    ADD CONSTRAINT fk_quotations_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: quotations fk_quotations_customer; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.quotations
    ADD CONSTRAINT fk_quotations_customer FOREIGN KEY (customer_id) REFERENCES softivite.customers(id);


--
-- Name: quotation_items fk_quotations_items; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.quotation_items
    ADD CONSTRAINT fk_quotations_items FOREIGN KEY (quotation_id) REFERENCES softivite.quotations(id);


--
-- Name: quotations fk_quotations_reviewed_by; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.quotations
    ADD CONSTRAINT fk_quotations_reviewed_by FOREIGN KEY (reviewed_by_id) REFERENCES public.users(id);


--
-- Name: return_items fk_return_items_product; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.return_items
    ADD CONSTRAINT fk_return_items_product FOREIGN KEY (product_id) REFERENCES softivite.products(id);


--
-- Name: returns fk_returns_approved_by; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.returns
    ADD CONSTRAINT fk_returns_approved_by FOREIGN KEY (approved_by_id) REFERENCES public.users(id);


--
-- Name: returns fk_returns_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.returns
    ADD CONSTRAINT fk_returns_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: returns fk_returns_created_by; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.returns
    ADD CONSTRAINT fk_returns_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: returns fk_returns_customer; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.returns
    ADD CONSTRAINT fk_returns_customer FOREIGN KEY (customer_id) REFERENCES softivite.customers(id);


--
-- Name: return_items fk_returns_items; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.return_items
    ADD CONSTRAINT fk_returns_items FOREIGN KEY (return_id) REFERENCES softivite.returns(id);


--
-- Name: returns fk_returns_order; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.returns
    ADD CONSTRAINT fk_returns_order FOREIGN KEY (order_id) REFERENCES softivite.orders(id);


--
-- Name: services fk_services_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.services
    ADD CONSTRAINT fk_services_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: services fk_services_category; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.services
    ADD CONSTRAINT fk_services_category FOREIGN KEY (category_id) REFERENCES softivite.categories(id);


--
-- Name: shifts fk_shifts_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.shifts
    ADD CONSTRAINT fk_shifts_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: split_bill_groups fk_split_bill_groups_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.split_bill_groups
    ADD CONSTRAINT fk_split_bill_groups_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: split_bill_groups fk_split_bill_groups_original_order; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.split_bill_groups
    ADD CONSTRAINT fk_split_bill_groups_original_order FOREIGN KEY (original_order_id) REFERENCES softivite.orders(id);


--
-- Name: split_bill_groups fk_split_bill_groups_table; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.split_bill_groups
    ADD CONSTRAINT fk_split_bill_groups_table FOREIGN KEY (table_id) REFERENCES softivite.dining_tables(id);


--
-- Name: stock_batches fk_stock_batches_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_batches
    ADD CONSTRAINT fk_stock_batches_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: stock_batches fk_stock_batches_product; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_batches
    ADD CONSTRAINT fk_stock_batches_product FOREIGN KEY (product_id) REFERENCES softivite.products(id);


--
-- Name: stock_movements fk_stock_movements_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_movements
    ADD CONSTRAINT fk_stock_movements_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: stock_movements fk_stock_movements_tenant; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_movements
    ADD CONSTRAINT fk_stock_movements_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: stock_transfer_items fk_stock_transfer_items_product; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_transfer_items
    ADD CONSTRAINT fk_stock_transfer_items_product FOREIGN KEY (product_id) REFERENCES softivite.products(id);


--
-- Name: stock_transfers fk_stock_transfers_created_by; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_transfers
    ADD CONSTRAINT fk_stock_transfers_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: stock_transfer_items fk_stock_transfers_items; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stock_transfer_items
    ADD CONSTRAINT fk_stock_transfers_items FOREIGN KEY (transfer_id) REFERENCES softivite.stock_transfers(id);


--
-- Name: stocktake_entries fk_stocktake_entries_product; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stocktake_entries
    ADD CONSTRAINT fk_stocktake_entries_product FOREIGN KEY (product_id) REFERENCES softivite.products(id);


--
-- Name: stocktake_sessions fk_stocktake_sessions_branch; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stocktake_sessions
    ADD CONSTRAINT fk_stocktake_sessions_branch FOREIGN KEY (branch_id) REFERENCES softivite.branches(id);


--
-- Name: stocktake_sessions fk_stocktake_sessions_created_by; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stocktake_sessions
    ADD CONSTRAINT fk_stocktake_sessions_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: stocktake_entries fk_stocktake_sessions_entries; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.stocktake_entries
    ADD CONSTRAINT fk_stocktake_sessions_entries FOREIGN KEY (session_id) REFERENCES softivite.stocktake_sessions(id);


--
-- Name: supplier_ledger_entries fk_supplier_ledger_entries_supplier; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.supplier_ledger_entries
    ADD CONSTRAINT fk_supplier_ledger_entries_supplier FOREIGN KEY (supplier_id) REFERENCES softivite.suppliers(id);


--
-- Name: supplier_products fk_supplier_products_product; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.supplier_products
    ADD CONSTRAINT fk_supplier_products_product FOREIGN KEY (product_id) REFERENCES softivite.products(id);


--
-- Name: supplier_products fk_supplier_products_supplier; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.supplier_products
    ADD CONSTRAINT fk_supplier_products_supplier FOREIGN KEY (supplier_id) REFERENCES softivite.suppliers(id);


--
-- Name: supplier_profiles fk_supplier_profiles_supplier; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.supplier_profiles
    ADD CONSTRAINT fk_supplier_profiles_supplier FOREIGN KEY (supplier_id) REFERENCES softivite.suppliers(id);


--
-- Name: supplier_profiles fk_supplier_profiles_user; Type: FK CONSTRAINT; Schema: softivite; Owner: -
--

ALTER TABLE ONLY softivite.supplier_profiles
    ADD CONSTRAINT fk_supplier_profiles_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: api_keys fk_api_keys_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.api_keys
    ADD CONSTRAINT fk_api_keys_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: api_request_logs fk_api_request_logs_tenant; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.api_request_logs
    ADD CONSTRAINT fk_api_request_logs_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: appointments fk_appointments_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.appointments
    ADD CONSTRAINT fk_appointments_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: attendances fk_attendances_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.attendances
    ADD CONSTRAINT fk_attendances_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: audit_logs fk_audit_logs_tenant; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.audit_logs
    ADD CONSTRAINT fk_audit_logs_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: audit_logs fk_audit_logs_user; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.audit_logs
    ADD CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: branches fk_branches_tenant; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.branches
    ADD CONSTRAINT fk_branches_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: cash_drawer_sessions fk_cash_drawer_sessions_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.cash_drawer_sessions
    ADD CONSTRAINT fk_cash_drawer_sessions_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: customers fk_customers_tier; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.customers
    ADD CONSTRAINT fk_customers_tier FOREIGN KEY (tier_id) REFERENCES thinkce.customer_tiers(id);


--
-- Name: delivery_orders fk_delivery_orders_driver; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.delivery_orders
    ADD CONSTRAINT fk_delivery_orders_driver FOREIGN KEY (driver_id) REFERENCES thinkce.delivery_drivers(id);


--
-- Name: delivery_orders fk_delivery_orders_order; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.delivery_orders
    ADD CONSTRAINT fk_delivery_orders_order FOREIGN KEY (order_id) REFERENCES thinkce.orders(id);


--
-- Name: dining_tables fk_dining_tables_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.dining_tables
    ADD CONSTRAINT fk_dining_tables_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: expenses fk_expenses_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.expenses
    ADD CONSTRAINT fk_expenses_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: expenses fk_expenses_category; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.expenses
    ADD CONSTRAINT fk_expenses_category FOREIGN KEY (category_id) REFERENCES thinkce.categories(id);


--
-- Name: expenses fk_expenses_created_by; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.expenses
    ADD CONSTRAINT fk_expenses_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: ledger_lines fk_journal_entries_lines; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.ledger_lines
    ADD CONSTRAINT fk_journal_entries_lines FOREIGN KEY (journal_entry_id) REFERENCES thinkce.journal_entries(id);


--
-- Name: kds_tickets fk_kds_tickets_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.kds_tickets
    ADD CONSTRAINT fk_kds_tickets_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: kds_tickets fk_kds_tickets_order; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.kds_tickets
    ADD CONSTRAINT fk_kds_tickets_order FOREIGN KEY (order_id) REFERENCES thinkce.orders(id);


--
-- Name: kds_tickets fk_kds_tickets_table; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.kds_tickets
    ADD CONSTRAINT fk_kds_tickets_table FOREIGN KEY (table_id) REFERENCES thinkce.dining_tables(id);


--
-- Name: leave_requests fk_leave_requests_reviewed_by; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.leave_requests
    ADD CONSTRAINT fk_leave_requests_reviewed_by FOREIGN KEY (reviewed_by_id) REFERENCES public.users(id);


--
-- Name: ledger_lines fk_ledger_lines_account; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.ledger_lines
    ADD CONSTRAINT fk_ledger_lines_account FOREIGN KEY (account_id) REFERENCES thinkce.ledger_accounts(id);


--
-- Name: loyalty_transactions fk_loyalty_transactions_tenant; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.loyalty_transactions
    ADD CONSTRAINT fk_loyalty_transactions_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: notifications fk_notifications_user; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.notifications
    ADD CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: order_items fk_order_items_product; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.order_items
    ADD CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES thinkce.products(id);


--
-- Name: orders fk_orders_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.orders
    ADD CONSTRAINT fk_orders_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: orders fk_orders_cashier; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.orders
    ADD CONSTRAINT fk_orders_cashier FOREIGN KEY (cashier_id) REFERENCES public.users(id);


--
-- Name: orders fk_orders_customer; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.orders
    ADD CONSTRAINT fk_orders_customer FOREIGN KEY (customer_id) REFERENCES thinkce.customers(id);


--
-- Name: order_items fk_orders_items; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.order_items
    ADD CONSTRAINT fk_orders_items FOREIGN KEY (order_id) REFERENCES thinkce.orders(id);


--
-- Name: print_jobs fk_print_jobs_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.print_jobs
    ADD CONSTRAINT fk_print_jobs_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: products fk_products_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.products
    ADD CONSTRAINT fk_products_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: products fk_products_category; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.products
    ADD CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES thinkce.categories(id);


--
-- Name: products fk_products_supplier; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.products
    ADD CONSTRAINT fk_products_supplier FOREIGN KEY (supplier_id) REFERENCES thinkce.suppliers(id);


--
-- Name: domains fk_public_tenants_domains; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.domains
    ADD CONSTRAINT fk_public_tenants_domains FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: purchase_order_items fk_purchase_order_items_product; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.purchase_order_items
    ADD CONSTRAINT fk_purchase_order_items_product FOREIGN KEY (product_id) REFERENCES thinkce.products(id);


--
-- Name: purchase_orders fk_purchase_orders_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.purchase_orders
    ADD CONSTRAINT fk_purchase_orders_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: purchase_order_items fk_purchase_orders_items; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.purchase_order_items
    ADD CONSTRAINT fk_purchase_orders_items FOREIGN KEY (po_id) REFERENCES thinkce.purchase_orders(id);


--
-- Name: purchase_orders fk_purchase_orders_supplier; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.purchase_orders
    ADD CONSTRAINT fk_purchase_orders_supplier FOREIGN KEY (supplier_id) REFERENCES thinkce.suppliers(id);


--
-- Name: quotations fk_quotations_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.quotations
    ADD CONSTRAINT fk_quotations_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: quotations fk_quotations_created_by; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.quotations
    ADD CONSTRAINT fk_quotations_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: quotations fk_quotations_customer; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.quotations
    ADD CONSTRAINT fk_quotations_customer FOREIGN KEY (customer_id) REFERENCES thinkce.customers(id);


--
-- Name: quotation_items fk_quotations_items; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.quotation_items
    ADD CONSTRAINT fk_quotations_items FOREIGN KEY (quotation_id) REFERENCES thinkce.quotations(id);


--
-- Name: quotations fk_quotations_reviewed_by; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.quotations
    ADD CONSTRAINT fk_quotations_reviewed_by FOREIGN KEY (reviewed_by_id) REFERENCES public.users(id);


--
-- Name: return_items fk_return_items_product; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.return_items
    ADD CONSTRAINT fk_return_items_product FOREIGN KEY (product_id) REFERENCES thinkce.products(id);


--
-- Name: returns fk_returns_approved_by; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.returns
    ADD CONSTRAINT fk_returns_approved_by FOREIGN KEY (approved_by_id) REFERENCES public.users(id);


--
-- Name: returns fk_returns_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.returns
    ADD CONSTRAINT fk_returns_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: returns fk_returns_created_by; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.returns
    ADD CONSTRAINT fk_returns_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: returns fk_returns_customer; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.returns
    ADD CONSTRAINT fk_returns_customer FOREIGN KEY (customer_id) REFERENCES thinkce.customers(id);


--
-- Name: return_items fk_returns_items; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.return_items
    ADD CONSTRAINT fk_returns_items FOREIGN KEY (return_id) REFERENCES thinkce.returns(id);


--
-- Name: returns fk_returns_order; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.returns
    ADD CONSTRAINT fk_returns_order FOREIGN KEY (order_id) REFERENCES thinkce.orders(id);


--
-- Name: services fk_services_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.services
    ADD CONSTRAINT fk_services_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: services fk_services_category; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.services
    ADD CONSTRAINT fk_services_category FOREIGN KEY (category_id) REFERENCES thinkce.categories(id);


--
-- Name: shifts fk_shifts_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.shifts
    ADD CONSTRAINT fk_shifts_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: split_bill_groups fk_split_bill_groups_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.split_bill_groups
    ADD CONSTRAINT fk_split_bill_groups_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: split_bill_groups fk_split_bill_groups_original_order; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.split_bill_groups
    ADD CONSTRAINT fk_split_bill_groups_original_order FOREIGN KEY (original_order_id) REFERENCES thinkce.orders(id);


--
-- Name: split_bill_groups fk_split_bill_groups_table; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.split_bill_groups
    ADD CONSTRAINT fk_split_bill_groups_table FOREIGN KEY (table_id) REFERENCES thinkce.dining_tables(id);


--
-- Name: stock_batches fk_stock_batches_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_batches
    ADD CONSTRAINT fk_stock_batches_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: stock_batches fk_stock_batches_product; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_batches
    ADD CONSTRAINT fk_stock_batches_product FOREIGN KEY (product_id) REFERENCES thinkce.products(id);


--
-- Name: stock_movements fk_stock_movements_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_movements
    ADD CONSTRAINT fk_stock_movements_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: stock_movements fk_stock_movements_tenant; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_movements
    ADD CONSTRAINT fk_stock_movements_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- Name: stock_transfer_items fk_stock_transfer_items_product; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_transfer_items
    ADD CONSTRAINT fk_stock_transfer_items_product FOREIGN KEY (product_id) REFERENCES thinkce.products(id);


--
-- Name: stock_transfers fk_stock_transfers_created_by; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_transfers
    ADD CONSTRAINT fk_stock_transfers_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: stock_transfer_items fk_stock_transfers_items; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stock_transfer_items
    ADD CONSTRAINT fk_stock_transfers_items FOREIGN KEY (transfer_id) REFERENCES thinkce.stock_transfers(id);


--
-- Name: stocktake_entries fk_stocktake_entries_product; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stocktake_entries
    ADD CONSTRAINT fk_stocktake_entries_product FOREIGN KEY (product_id) REFERENCES thinkce.products(id);


--
-- Name: stocktake_sessions fk_stocktake_sessions_branch; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stocktake_sessions
    ADD CONSTRAINT fk_stocktake_sessions_branch FOREIGN KEY (branch_id) REFERENCES thinkce.branches(id);


--
-- Name: stocktake_sessions fk_stocktake_sessions_created_by; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stocktake_sessions
    ADD CONSTRAINT fk_stocktake_sessions_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: stocktake_entries fk_stocktake_sessions_entries; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.stocktake_entries
    ADD CONSTRAINT fk_stocktake_sessions_entries FOREIGN KEY (session_id) REFERENCES thinkce.stocktake_sessions(id);


--
-- Name: supplier_ledger_entries fk_supplier_ledger_entries_supplier; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.supplier_ledger_entries
    ADD CONSTRAINT fk_supplier_ledger_entries_supplier FOREIGN KEY (supplier_id) REFERENCES thinkce.suppliers(id);


--
-- Name: supplier_products fk_supplier_products_product; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.supplier_products
    ADD CONSTRAINT fk_supplier_products_product FOREIGN KEY (product_id) REFERENCES thinkce.products(id);


--
-- Name: supplier_products fk_supplier_products_supplier; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.supplier_products
    ADD CONSTRAINT fk_supplier_products_supplier FOREIGN KEY (supplier_id) REFERENCES thinkce.suppliers(id);


--
-- Name: supplier_profiles fk_supplier_profiles_supplier; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.supplier_profiles
    ADD CONSTRAINT fk_supplier_profiles_supplier FOREIGN KEY (supplier_id) REFERENCES thinkce.suppliers(id);


--
-- Name: supplier_profiles fk_supplier_profiles_user; Type: FK CONSTRAINT; Schema: thinkce; Owner: -
--

ALTER TABLE ONLY thinkce.supplier_profiles
    ADD CONSTRAINT fk_supplier_profiles_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

\unrestrict x2dRdwLoOXbQaHuj4FuoPoUThxvqdiJEfId3pWWIP2BF3RhJYodOpDxksQ178CY

