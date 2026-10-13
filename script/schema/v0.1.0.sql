--
-- PostgreSQL database dump
--

\restrict h6vzkawMdZNbL8gFGgtHHg4bP3CY5jLXjjlkYCRaTr65qQ4RDij4rlMQ1k5jo0J

-- Dumped from database version 16.15
-- Dumped by pg_dump version 16.15

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA IF NOT EXISTS public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: system_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_config (
    id bigint NOT NULL,
    config_name character varying(64) NOT NULL,
    config_key character varying(128) NOT NULL,
    config_value text DEFAULT ''::text NOT NULL,
    config_type character varying(16) DEFAULT 'system'::character varying NOT NULL,
    remark character varying(255) DEFAULT ''::character varying NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: system_config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_config_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_config_id_seq OWNED BY public.system_config.id;


--
-- Name: system_dept; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_dept (
    id bigint NOT NULL,
    parent_id bigint DEFAULT 0 NOT NULL,
    ancestors character varying(500) DEFAULT '0'::character varying NOT NULL,
    dept_name character varying(64) NOT NULL,
    order_num integer DEFAULT 0 NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    leader_user_id bigint DEFAULT 0 NOT NULL
);


--
-- Name: TABLE system_dept; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_dept IS '部门';


--
-- Name: system_dept_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_dept_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_dept_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_dept_id_seq OWNED BY public.system_dept.id;


--
-- Name: system_dict_data; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_dict_data (
    id bigint NOT NULL,
    dict_type character varying(64) NOT NULL,
    dict_label character varying(64) NOT NULL,
    dict_value character varying(64) NOT NULL,
    css_class character varying(64) DEFAULT ''::character varying NOT NULL,
    list_order integer DEFAULT 0 NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    is_default smallint DEFAULT 0 NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    theme character varying(16) DEFAULT ''::character varying NOT NULL,
    color character varying(32) DEFAULT ''::character varying NOT NULL
);


--
-- Name: system_dict_data_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_dict_data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_dict_data_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_dict_data_id_seq OWNED BY public.system_dict_data.id;


--
-- Name: system_dict_group; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_dict_group (
    id bigint NOT NULL,
    group_name character varying(64) NOT NULL,
    order_num integer DEFAULT 0 NOT NULL,
    remark character varying(255) DEFAULT ''::character varying NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: system_dict_group_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_dict_group_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_dict_group_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_dict_group_id_seq OWNED BY public.system_dict_group.id;


--
-- Name: system_dict_type; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_dict_type (
    id bigint NOT NULL,
    dict_name character varying(64) NOT NULL,
    dict_type character varying(64) NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    order_num integer DEFAULT 0 NOT NULL,
    remark character varying(255) DEFAULT ''::character varying NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    group_id bigint DEFAULT 0 NOT NULL
);


--
-- Name: system_dict_type_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_dict_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_dict_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_dict_type_id_seq OWNED BY public.system_dict_type.id;


--
-- Name: system_error_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_error_log (
    id bigint NOT NULL,
    trace_id character varying(64) DEFAULT ''::character varying NOT NULL,
    user_id bigint DEFAULT 0 NOT NULL,
    user_name character varying(64) DEFAULT ''::character varying NOT NULL,
    err_type smallint DEFAULT 1 NOT NULL,
    err_level smallint DEFAULT 1 NOT NULL,
    handle_status smallint DEFAULT 0 NOT NULL,
    err_code integer DEFAULT 0 NOT NULL,
    err_msg text DEFAULT ''::text NOT NULL,
    stack text DEFAULT ''::text NOT NULL,
    request_uri character varying(255) DEFAULT ''::character varying NOT NULL,
    http_method character varying(16) DEFAULT ''::character varying NOT NULL,
    module character varying(64) DEFAULT ''::character varying NOT NULL,
    request_headers text DEFAULT ''::text NOT NULL,
    request_body text DEFAULT ''::text NOT NULL,
    cost_ms integer DEFAULT 0 NOT NULL,
    client_ip character varying(64) DEFAULT ''::character varying NOT NULL,
    handled_at timestamp without time zone,
    handle_remark character varying(512) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_error_log; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_error_log IS '系统异常日志';


--
-- Name: system_error_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_error_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_error_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_error_log_id_seq OWNED BY public.system_error_log.id;


--
-- Name: system_job; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_job (
    id bigint NOT NULL,
    job_key character varying(128) NOT NULL,
    job_name character varying(255) NOT NULL,
    cron character varying(128) NOT NULL,
    timeout_sec integer DEFAULT 60 NOT NULL,
    params_json text DEFAULT '{}'::text NOT NULL,
    retry_count integer DEFAULT 0 NOT NULL,
    status smallint DEFAULT 1 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    active_run_token character varying(64) DEFAULT ''::character varying NOT NULL,
    config_version integer DEFAULT 1 NOT NULL,
    max_instances integer DEFAULT 1 NOT NULL,
    active_run_count integer DEFAULT 0 NOT NULL,
    CONSTRAINT ck_system_job_active_run_count CHECK ((active_run_count >= 0)),
    CONSTRAINT ck_system_job_max_instances CHECK ((max_instances > 0)),
    CONSTRAINT ck_system_job_retry CHECK ((retry_count >= 0)),
    CONSTRAINT ck_system_job_status CHECK ((status = ANY (ARRAY[0, 1]))),
    CONSTRAINT ck_system_job_timeout CHECK ((timeout_sec > 0))
);


--
-- Name: system_job_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_job_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_job_id_seq OWNED BY public.system_job.id;


--
-- Name: system_job_run; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_job_run (
    id bigint NOT NULL,
    job_id bigint NOT NULL,
    job_key character varying(128) NOT NULL,
    trigger_type character varying(20) NOT NULL,
    attempt integer NOT NULL,
    status character varying(32) NOT NULL,
    triggered_at timestamp without time zone NOT NULL,
    finished_at timestamp without time zone,
    output text DEFAULT ''::text NOT NULL,
    error_message text DEFAULT ''::text NOT NULL,
    job_name character varying(128) DEFAULT ''::character varying NOT NULL,
    params_json text DEFAULT '{}'::text NOT NULL,
    timeout_sec integer DEFAULT 0 NOT NULL,
    config_version integer DEFAULT 1 NOT NULL
);


--
-- Name: system_job_run_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_job_run_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_job_run_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_job_run_id_seq OWNED BY public.system_job_run.id;


--
-- Name: system_job_run_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_job_run_log (
    id bigint NOT NULL,
    run_id bigint NOT NULL,
    job_key character varying(128) NOT NULL,
    level character varying(16) NOT NULL,
    stage character varying(32) NOT NULL,
    message text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: system_job_run_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_job_run_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_job_run_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_job_run_log_id_seq OWNED BY public.system_job_run_log.id;


--
-- Name: system_login_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_login_log (
    id bigint NOT NULL,
    trace_id character varying(64) DEFAULT ''::character varying NOT NULL,
    user_id bigint DEFAULT 0 NOT NULL,
    user_name character varying(64) DEFAULT ''::character varying NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    fail_reason character varying(255) DEFAULT ''::character varying NOT NULL,
    client_ip character varying(64) DEFAULT ''::character varying NOT NULL,
    user_agent character varying(512) DEFAULT ''::character varying NOT NULL,
    browser character varying(64) DEFAULT ''::character varying NOT NULL,
    os character varying(64) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_login_log; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_login_log IS '登录日志';


--
-- Name: system_login_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_login_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_login_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_login_log_id_seq OWNED BY public.system_login_log.id;


--
-- Name: system_menu; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_menu (
    id bigint NOT NULL,
    parent_id bigint NOT NULL,
    permission_code character varying(255) NOT NULL,
    permission_type character varying(255),
    menu_type character varying(255),
    group_name character varying(36),
    view_mode character varying(255),
    title character varying(255),
    path character varying(255),
    query_param character varying(255),
    sort integer,
    icon character varying(255),
    state character varying(255),
    description character varying(255),
    tag text,
    component character varying(512),
    visible character varying(255),
    is_cache character varying(255),
    is_remove smallint,
    update_time timestamp(6) without time zone,
    create_time timestamp(6) without time zone,
    create_by character varying(36),
    update_by character varying(36),
    redirect character varying(512) DEFAULT ''::character varying NOT NULL,
    expanded boolean DEFAULT false NOT NULL,
    hidden_breadcrumb boolean DEFAULT false NOT NULL,
    perms character varying(255) DEFAULT ''::character varying NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    action_type character varying(32) DEFAULT ''::character varying NOT NULL
);


--
-- Name: TABLE system_menu; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_menu IS '系统菜单';


--
-- Name: COLUMN system_menu.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.id IS '主键ID';


--
-- Name: COLUMN system_menu.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.parent_id IS '上级ID';


--
-- Name: COLUMN system_menu.permission_code; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.permission_code IS '权限编号';


--
-- Name: COLUMN system_menu.permission_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.permission_type IS '权限类型';


--
-- Name: COLUMN system_menu.menu_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.menu_type IS '菜单类型';


--
-- Name: COLUMN system_menu.group_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.group_name IS '菜单分组';


--
-- Name: COLUMN system_menu.view_mode; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.view_mode IS '显示模式';


--
-- Name: COLUMN system_menu.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.title IS '菜单标题';


--
-- Name: COLUMN system_menu.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.path IS '菜单路径';


--
-- Name: COLUMN system_menu.query_param; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.query_param IS '路由参数';


--
-- Name: COLUMN system_menu.sort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.sort IS '菜单排序';


--
-- Name: COLUMN system_menu.icon; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.icon IS '菜单图标';


--
-- Name: COLUMN system_menu.state; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.state IS '菜单状态';


--
-- Name: COLUMN system_menu.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.description IS '菜单描述';


--
-- Name: COLUMN system_menu.tag; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.tag IS '菜单标签';


--
-- Name: COLUMN system_menu.component; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.component IS '组件地址';


--
-- Name: COLUMN system_menu.visible; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.visible IS '是否可见：0-隐藏, 1-显示';


--
-- Name: COLUMN system_menu.is_cache; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.is_cache IS '开启缓存：0-开启，1-关闭';


--
-- Name: COLUMN system_menu.is_remove; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.is_remove IS '是否删除：0-未删除，1-已删除';


--
-- Name: COLUMN system_menu.update_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.update_time IS '修改日期';


--
-- Name: COLUMN system_menu.create_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.create_time IS '创建日期';


--
-- Name: COLUMN system_menu.create_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.create_by IS '创建人ID';


--
-- Name: COLUMN system_menu.update_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.update_by IS '修改人ID';


--
-- Name: COLUMN system_menu.redirect; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.redirect IS '重定向路径';


--
-- Name: COLUMN system_menu.expanded; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.expanded IS '菜单是否默认展开';


--
-- Name: COLUMN system_menu.hidden_breadcrumb; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.hidden_breadcrumb IS '面包屑中是否隐藏';


--
-- Name: COLUMN system_menu.perms; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.perms IS '权限标识 e.g. system:user:list';


--
-- Name: COLUMN system_menu.action_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_menu.action_type IS '按钮操作类型（sys_action_type），仅 menu_type=F 使用';


--
-- Name: system_message_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_message_config (
    id bigint NOT NULL,
    title character varying(255) NOT NULL,
    message_type character varying(20) NOT NULL,
    supplier_type character varying(20) NOT NULL,
    config_json text,
    status smallint DEFAULT 1 NOT NULL,
    remark character varying(500) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_message_config; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_message_config IS '消息配置表';


--
-- Name: system_message_config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_message_config_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_message_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_message_config_id_seq OWNED BY public.system_message_config.id;


--
-- Name: system_message_key; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_message_key (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    message_key character varying(50) NOT NULL,
    remark character varying(500) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_message_key; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_message_key IS '消息常量表';


--
-- Name: system_message_key_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_message_key_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_message_key_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_message_key_id_seq OWNED BY public.system_message_key.id;


--
-- Name: system_message_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_message_log (
    id bigint NOT NULL,
    message_template_id bigint,
    message_key character varying(50) NOT NULL,
    message_template_name character varying(255) DEFAULT ''::character varying NOT NULL,
    message_type character varying(20) NOT NULL,
    template_mode character varying(20) DEFAULT ''::character varying NOT NULL,
    account character varying(255) NOT NULL,
    title character varying(255) DEFAULT ''::character varying NOT NULL,
    template_id character varying(100) DEFAULT ''::character varying NOT NULL,
    content text,
    message_config_title character varying(255) DEFAULT ''::character varying NOT NULL,
    supplier_type character varying(20) DEFAULT ''::character varying NOT NULL,
    is_success smallint,
    response_body character varying(1000),
    log_time timestamp without time zone NOT NULL,
    cost_time bigint
);


--
-- Name: TABLE system_message_log; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_message_log IS '消息发送记录表';


--
-- Name: system_message_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_message_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_message_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_message_log_id_seq OWNED BY public.system_message_log.id;


--
-- Name: system_message_template; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_message_template (
    id bigint NOT NULL,
    template_name character varying(255) NOT NULL,
    message_config_id bigint NOT NULL,
    message_key_id bigint NOT NULL,
    message_key character varying(50) NOT NULL,
    message_type character varying(20) NOT NULL,
    template_mode character varying(20) NOT NULL,
    title character varying(255) DEFAULT ''::character varying NOT NULL,
    signature character varying(100) DEFAULT ''::character varying NOT NULL,
    template_id character varying(100) DEFAULT ''::character varying NOT NULL,
    content text,
    vars_json text,
    status smallint DEFAULT 1 NOT NULL,
    remark character varying(500) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_message_template; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_message_template IS '消息模板表';


--
-- Name: system_message_template_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_message_template_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_message_template_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_message_template_id_seq OWNED BY public.system_message_template.id;


--
-- Name: system_notice; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_notice (
    id bigint NOT NULL,
    title character varying(128) NOT NULL,
    notice_type smallint DEFAULT 1 NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    publisher character varying(64) DEFAULT ''::character varying NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_notice; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_notice IS '通知公告';


--
-- Name: system_notice_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_notice_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_notice_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_notice_id_seq OWNED BY public.system_notice.id;


--
-- Name: system_notice_read; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_notice_read (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    notice_id bigint NOT NULL,
    read_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_notice_read; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_notice_read IS '通知已读记录（消费侧）';


--
-- Name: system_notice_read_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_notice_read_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_notice_read_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_notice_read_id_seq OWNED BY public.system_notice_read.id;


--
-- Name: system_oper_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_oper_log (
    id bigint NOT NULL,
    trace_id character varying(64) DEFAULT ''::character varying NOT NULL,
    user_id bigint DEFAULT 0 NOT NULL,
    user_name character varying(64) DEFAULT ''::character varying NOT NULL,
    oper_type smallint DEFAULT 0 NOT NULL,
    request_uri character varying(255) DEFAULT ''::character varying NOT NULL,
    http_method character varying(16) DEFAULT ''::character varying NOT NULL,
    module character varying(64) DEFAULT ''::character varying NOT NULL,
    oper_status smallint DEFAULT 0 NOT NULL,
    request_headers text DEFAULT ''::text NOT NULL,
    request_body text DEFAULT ''::text NOT NULL,
    response_body text DEFAULT ''::text NOT NULL,
    error_msg text DEFAULT ''::text NOT NULL,
    remark character varying(512) DEFAULT ''::character varying NOT NULL,
    cost_ms integer DEFAULT 0 NOT NULL,
    client_ip character varying(64) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_oper_log; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_oper_log IS '操作日志';


--
-- Name: system_oper_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_oper_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_oper_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_oper_log_id_seq OWNED BY public.system_oper_log.id;


--
-- Name: system_oss; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_oss (
    id bigint NOT NULL,
    file_name character varying(255) NOT NULL,
    original_name character varying(255) NOT NULL,
    file_suffix character varying(32) DEFAULT ''::character varying NOT NULL,
    url character varying(500) NOT NULL,
    size bigint DEFAULT 0 NOT NULL,
    content_type character varying(128) DEFAULT ''::character varying NOT NULL,
    is_lock boolean DEFAULT false NOT NULL,
    service character varying(64) DEFAULT ''::character varying NOT NULL,
    create_by character varying(64) DEFAULT ''::character varying NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_oss; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_oss IS 'OSS文件台账';


--
-- Name: system_oss_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_oss_config (
    id bigint NOT NULL,
    config_key character varying(64) NOT NULL,
    access_key character varying(255) DEFAULT ''::character varying NOT NULL,
    secret_key character varying(255) DEFAULT ''::character varying NOT NULL,
    bucket_name character varying(255) NOT NULL,
    prefix character varying(255) DEFAULT ''::character varying NOT NULL,
    endpoint character varying(255) NOT NULL,
    domain character varying(255) DEFAULT ''::character varying NOT NULL,
    is_https character(1) DEFAULT 'N'::bpchar NOT NULL,
    region character varying(64) DEFAULT ''::character varying NOT NULL,
    access_policy character(1) DEFAULT '1'::bpchar NOT NULL,
    status character(1) DEFAULT '0'::bpchar NOT NULL,
    ext1 character varying(255) DEFAULT ''::character varying NOT NULL,
    remark character varying(500) DEFAULT ''::character varying NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_oss_config; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_oss_config IS '对象存储配置';


--
-- Name: system_oss_config_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_oss_config_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_oss_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_oss_config_id_seq OWNED BY public.system_oss_config.id;


--
-- Name: system_oss_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_oss_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_oss_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_oss_id_seq OWNED BY public.system_oss.id;


--
-- Name: system_oss_rule; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_oss_rule (
    id bigint NOT NULL,
    rule_name character varying(255) NOT NULL,
    domain character varying(255) NOT NULL,
    mime_type character varying(255) DEFAULT ''::character varying NOT NULL,
    rule text NOT NULL,
    is_overwrite character(1) DEFAULT 'N'::bpchar NOT NULL,
    is_default character(1) DEFAULT 'N'::bpchar NOT NULL,
    status character(1) DEFAULT '1'::bpchar NOT NULL,
    rule_sort integer DEFAULT 0 NOT NULL,
    remark character varying(500) DEFAULT ''::character varying NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_oss_rule; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_oss_rule IS 'OSS处理规则';


--
-- Name: system_oss_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_oss_rule_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_oss_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_oss_rule_id_seq OWNED BY public.system_oss_rule.id;


--
-- Name: system_post; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_post (
    id bigint NOT NULL,
    post_code character varying(64) NOT NULL,
    post_name character varying(64) NOT NULL,
    order_num integer DEFAULT 0 NOT NULL,
    status smallint DEFAULT 0 NOT NULL,
    del_flag smallint DEFAULT 0 NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: TABLE system_post; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_post IS '岗位';


--
-- Name: system_post_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_post_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_post_id_seq OWNED BY public.system_post.id;


--
-- Name: system_role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_role (
    id bigint NOT NULL,
    name character varying(36),
    code character varying(36),
    description character varying(510),
    state character varying(255),
    type character varying(255),
    is_remove smallint,
    updated_at timestamp(6) without time zone,
    created_at timestamp(6) without time zone,
    created_by bigint,
    updated_by bigint,
    deleted_at timestamp(6) without time zone,
    data_scope smallint DEFAULT 1 NOT NULL,
    order_num integer DEFAULT 0 NOT NULL
);


--
-- Name: TABLE system_role; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_role IS '系统角色';


--
-- Name: COLUMN system_role.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.id IS '主键';


--
-- Name: COLUMN system_role.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.name IS '角色名称';


--
-- Name: COLUMN system_role.code; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.code IS '角色编码';


--
-- Name: COLUMN system_role.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.description IS '角色描述';


--
-- Name: COLUMN system_role.state; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.state IS '角色状态';


--
-- Name: COLUMN system_role.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.type IS '是否内置';


--
-- Name: COLUMN system_role.is_remove; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.is_remove IS '是否删除';


--
-- Name: COLUMN system_role.updated_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.updated_at IS '修改日期';


--
-- Name: COLUMN system_role.created_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.created_at IS '创建日期';


--
-- Name: COLUMN system_role.created_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.created_by IS '创建人ID';


--
-- Name: COLUMN system_role.updated_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.updated_by IS '修改人ID';


--
-- Name: COLUMN system_role.deleted_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.deleted_at IS '删除日期';


--
-- Name: COLUMN system_role.data_scope; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role.data_scope IS '数据范围 1全部 2本部门 3本部门及以下 4仅本人 5自定义';


--
-- Name: system_role_dept; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_role_dept (
    id bigint NOT NULL,
    role_id bigint NOT NULL,
    dept_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: system_role_dept_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_role_dept_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_role_dept_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_role_dept_id_seq OWNED BY public.system_role_dept.id;


--
-- Name: system_role_menu; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_role_menu (
    id bigint NOT NULL,
    role_id bigint NOT NULL,
    menu_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: system_role_menu_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_role_menu_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_role_menu_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_role_menu_id_seq OWNED BY public.system_role_menu.id;


--
-- Name: system_role_permission; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_role_permission (
    id bigint NOT NULL,
    role_id bigint,
    permission_id bigint,
    is_remove smallint,
    update_time timestamp(6) without time zone,
    create_time timestamp(6) without time zone,
    create_by character varying(36),
    update_by character varying(36)
);


--
-- Name: TABLE system_role_permission; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_role_permission IS '系统角色权限关联表';


--
-- Name: COLUMN system_role_permission.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role_permission.id IS '主键';


--
-- Name: COLUMN system_role_permission.role_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role_permission.role_id IS '角色ID';


--
-- Name: COLUMN system_role_permission.permission_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role_permission.permission_id IS '权限ID';


--
-- Name: COLUMN system_role_permission.is_remove; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role_permission.is_remove IS '是否删除';


--
-- Name: COLUMN system_role_permission.update_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role_permission.update_time IS '修改日期';


--
-- Name: COLUMN system_role_permission.create_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role_permission.create_time IS '创建日期';


--
-- Name: COLUMN system_role_permission.create_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role_permission.create_by IS '创建人ID';


--
-- Name: COLUMN system_role_permission.update_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_role_permission.update_by IS '修改人ID';


--
-- Name: system_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public."system_user" (
    id bigint NOT NULL,
    account character varying(36),
    mobile_phone character varying(15),
    email character varying(255),
    username character varying(36),
    avatar_url character varying(255),
    state character varying(255),
    password character varying(64),
    is_remove smallint,
    update_time timestamp(6) without time zone,
    create_time timestamp(6) without time zone,
    create_by character varying(36),
    update_by character varying(36),
    dept_id bigint DEFAULT 0 NOT NULL,
    login_failure_count integer DEFAULT 0 NOT NULL,
    login_permanently_locked boolean DEFAULT false NOT NULL
);


--
-- Name: TABLE "system_user"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public."system_user" IS '系统用户';


--
-- Name: COLUMN "system_user".id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".id IS '用户唯一编码';


--
-- Name: COLUMN "system_user".account; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".account IS '账号';


--
-- Name: COLUMN "system_user".mobile_phone; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".mobile_phone IS '电话号码';


--
-- Name: COLUMN "system_user".email; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".email IS '邮箱地址';


--
-- Name: COLUMN "system_user".username; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".username IS '用户名称';


--
-- Name: COLUMN "system_user".avatar_url; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".avatar_url IS '用户头像';


--
-- Name: COLUMN "system_user".state; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".state IS '用户状态';


--
-- Name: COLUMN "system_user".password; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".password IS '用户密码';


--
-- Name: COLUMN "system_user".is_remove; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".is_remove IS '是否删除：0-未删除，1-已删除';


--
-- Name: COLUMN "system_user".update_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".update_time IS '修改时间';


--
-- Name: COLUMN "system_user".create_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".create_time IS '创建时间';


--
-- Name: COLUMN "system_user".create_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".create_by IS '创建人ID';


--
-- Name: COLUMN "system_user".update_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".update_by IS '修改人ID';


--
-- Name: COLUMN "system_user".dept_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public."system_user".dept_id IS '所属部门';


--
-- Name: system_user_acl_permission; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_user_acl_permission (
    id bigint NOT NULL,
    user_id character varying(36) NOT NULL,
    permission_id bigint NOT NULL,
    is_remove smallint,
    update_time timestamp(6) without time zone,
    create_time timestamp(6) without time zone,
    create_by character varying(36),
    update_by character varying(36)
);


--
-- Name: TABLE system_user_acl_permission; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_user_acl_permission IS '系统用户ACL权限';


--
-- Name: COLUMN system_user_acl_permission.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_acl_permission.id IS '主键';


--
-- Name: COLUMN system_user_acl_permission.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_acl_permission.user_id IS '用户ID';


--
-- Name: COLUMN system_user_acl_permission.permission_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_acl_permission.permission_id IS '权限ID';


--
-- Name: COLUMN system_user_acl_permission.is_remove; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_acl_permission.is_remove IS '是否删除';


--
-- Name: COLUMN system_user_acl_permission.update_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_acl_permission.update_time IS '修改日期';


--
-- Name: COLUMN system_user_acl_permission.create_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_acl_permission.create_time IS '创建日期';


--
-- Name: COLUMN system_user_acl_permission.create_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_acl_permission.create_by IS '创建人ID';


--
-- Name: COLUMN system_user_acl_permission.update_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_acl_permission.update_by IS '修改人ID';


--
-- Name: system_user_post; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_user_post (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    post_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: system_user_post_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_user_post_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_user_post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_user_post_id_seq OWNED BY public.system_user_post.id;


--
-- Name: system_user_role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_user_role (
    id bigint NOT NULL,
    user_id bigint,
    role_id bigint,
    type character varying(255),
    is_remove smallint,
    update_time timestamp(6) without time zone,
    create_time timestamp(6) without time zone,
    create_by character varying(36),
    update_by character varying(36)
);


--
-- Name: TABLE system_user_role; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.system_user_role IS '系统用户角色';


--
-- Name: COLUMN system_user_role.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.id IS '主键';


--
-- Name: COLUMN system_user_role.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.user_id IS '用户编号';


--
-- Name: COLUMN system_user_role.role_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.role_id IS '角色编号';


--
-- Name: COLUMN system_user_role.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.type IS '是否内置：0-内置，1-非内置';


--
-- Name: COLUMN system_user_role.is_remove; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.is_remove IS '是否删除：0-未删除，1-删除';


--
-- Name: COLUMN system_user_role.update_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.update_time IS '修改日期';


--
-- Name: COLUMN system_user_role.create_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.create_time IS '创建日期';


--
-- Name: COLUMN system_user_role.create_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.create_by IS '创建人ID';


--
-- Name: COLUMN system_user_role.update_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.system_user_role.update_by IS '修改人ID';


--
-- Name: system_config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_config ALTER COLUMN id SET DEFAULT nextval('public.system_config_id_seq'::regclass);


--
-- Name: system_dept id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_dept ALTER COLUMN id SET DEFAULT nextval('public.system_dept_id_seq'::regclass);


--
-- Name: system_dict_data id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_dict_data ALTER COLUMN id SET DEFAULT nextval('public.system_dict_data_id_seq'::regclass);


--
-- Name: system_dict_group id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_dict_group ALTER COLUMN id SET DEFAULT nextval('public.system_dict_group_id_seq'::regclass);


--
-- Name: system_dict_type id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_dict_type ALTER COLUMN id SET DEFAULT nextval('public.system_dict_type_id_seq'::regclass);


--
-- Name: system_error_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_error_log ALTER COLUMN id SET DEFAULT nextval('public.system_error_log_id_seq'::regclass);


--
-- Name: system_job id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_job ALTER COLUMN id SET DEFAULT nextval('public.system_job_id_seq'::regclass);


--
-- Name: system_job_run id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_job_run ALTER COLUMN id SET DEFAULT nextval('public.system_job_run_id_seq'::regclass);


--
-- Name: system_job_run_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_job_run_log ALTER COLUMN id SET DEFAULT nextval('public.system_job_run_log_id_seq'::regclass);


--
-- Name: system_login_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_login_log ALTER COLUMN id SET DEFAULT nextval('public.system_login_log_id_seq'::regclass);


--
-- Name: system_message_config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message_config ALTER COLUMN id SET DEFAULT nextval('public.system_message_config_id_seq'::regclass);


--
-- Name: system_message_key id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message_key ALTER COLUMN id SET DEFAULT nextval('public.system_message_key_id_seq'::regclass);


--
-- Name: system_message_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message_log ALTER COLUMN id SET DEFAULT nextval('public.system_message_log_id_seq'::regclass);


--
-- Name: system_message_template id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message_template ALTER COLUMN id SET DEFAULT nextval('public.system_message_template_id_seq'::regclass);


--
-- Name: system_notice id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_notice ALTER COLUMN id SET DEFAULT nextval('public.system_notice_id_seq'::regclass);


--
-- Name: system_notice_read id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_notice_read ALTER COLUMN id SET DEFAULT nextval('public.system_notice_read_id_seq'::regclass);


--
-- Name: system_oper_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oper_log ALTER COLUMN id SET DEFAULT nextval('public.system_oper_log_id_seq'::regclass);


--
-- Name: system_oss id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oss ALTER COLUMN id SET DEFAULT nextval('public.system_oss_id_seq'::regclass);


--
-- Name: system_oss_config id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oss_config ALTER COLUMN id SET DEFAULT nextval('public.system_oss_config_id_seq'::regclass);


--
-- Name: system_oss_rule id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oss_rule ALTER COLUMN id SET DEFAULT nextval('public.system_oss_rule_id_seq'::regclass);


--
-- Name: system_post id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_post ALTER COLUMN id SET DEFAULT nextval('public.system_post_id_seq'::regclass);


--
-- Name: system_role_dept id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_role_dept ALTER COLUMN id SET DEFAULT nextval('public.system_role_dept_id_seq'::regclass);


--
-- Name: system_role_menu id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_role_menu ALTER COLUMN id SET DEFAULT nextval('public.system_role_menu_id_seq'::regclass);


--
-- Name: system_user_post id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_user_post ALTER COLUMN id SET DEFAULT nextval('public.system_user_post_id_seq'::regclass);


--
-- Name: system_config system_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_config
    ADD CONSTRAINT system_config_pkey PRIMARY KEY (id);


--
-- Name: system_dept system_dept_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_dept
    ADD CONSTRAINT system_dept_pkey PRIMARY KEY (id);


--
-- Name: system_dict_data system_dict_data_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_dict_data
    ADD CONSTRAINT system_dict_data_pkey PRIMARY KEY (id);


--
-- Name: system_dict_group system_dict_group_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_dict_group
    ADD CONSTRAINT system_dict_group_pkey PRIMARY KEY (id);


--
-- Name: system_dict_type system_dict_type_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_dict_type
    ADD CONSTRAINT system_dict_type_pkey PRIMARY KEY (id);


--
-- Name: system_error_log system_error_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_error_log
    ADD CONSTRAINT system_error_log_pkey PRIMARY KEY (id);


--
-- Name: system_job system_job_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_job
    ADD CONSTRAINT system_job_pkey PRIMARY KEY (id);


--
-- Name: system_job_run_log system_job_run_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_job_run_log
    ADD CONSTRAINT system_job_run_log_pkey PRIMARY KEY (id);


--
-- Name: system_job_run system_job_run_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_job_run
    ADD CONSTRAINT system_job_run_pkey PRIMARY KEY (id);


--
-- Name: system_login_log system_login_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_login_log
    ADD CONSTRAINT system_login_log_pkey PRIMARY KEY (id);


--
-- Name: system_menu system_menu_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_menu
    ADD CONSTRAINT system_menu_pkey PRIMARY KEY (id);


--
-- Name: system_message_config system_message_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message_config
    ADD CONSTRAINT system_message_config_pkey PRIMARY KEY (id);


--
-- Name: system_message_key system_message_key_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message_key
    ADD CONSTRAINT system_message_key_pkey PRIMARY KEY (id);


--
-- Name: system_message_log system_message_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message_log
    ADD CONSTRAINT system_message_log_pkey PRIMARY KEY (id);


--
-- Name: system_message_template system_message_template_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_message_template
    ADD CONSTRAINT system_message_template_pkey PRIMARY KEY (id);


--
-- Name: system_notice system_notice_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_notice
    ADD CONSTRAINT system_notice_pkey PRIMARY KEY (id);


--
-- Name: system_notice_read system_notice_read_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_notice_read
    ADD CONSTRAINT system_notice_read_pkey PRIMARY KEY (id);


--
-- Name: system_oper_log system_oper_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oper_log
    ADD CONSTRAINT system_oper_log_pkey PRIMARY KEY (id);


--
-- Name: system_oss_config system_oss_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oss_config
    ADD CONSTRAINT system_oss_config_pkey PRIMARY KEY (id);


--
-- Name: system_oss system_oss_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oss
    ADD CONSTRAINT system_oss_pkey PRIMARY KEY (id);


--
-- Name: system_oss_rule system_oss_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oss_rule
    ADD CONSTRAINT system_oss_rule_pkey PRIMARY KEY (id);


--
-- Name: system_post system_post_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_post
    ADD CONSTRAINT system_post_pkey PRIMARY KEY (id);


--
-- Name: system_role_dept system_role_dept_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_role_dept
    ADD CONSTRAINT system_role_dept_pkey PRIMARY KEY (id);


--
-- Name: system_role_menu system_role_menu_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_role_menu
    ADD CONSTRAINT system_role_menu_pkey PRIMARY KEY (id);


--
-- Name: system_role_permission system_role_permission_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_role_permission
    ADD CONSTRAINT system_role_permission_pkey PRIMARY KEY (id);


--
-- Name: system_role system_role_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_role
    ADD CONSTRAINT system_role_pkey PRIMARY KEY (id);


--
-- Name: system_user system_user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."system_user"
    ADD CONSTRAINT system_user_pkey PRIMARY KEY (id);


--
-- Name: system_user_post system_user_post_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_user_post
    ADD CONSTRAINT system_user_post_pkey PRIMARY KEY (id);


--
-- Name: system_job uk_system_job_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_job
    ADD CONSTRAINT uk_system_job_key UNIQUE (job_key);


--
-- Name: system_oss_config uk_system_oss_config_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_oss_config
    ADD CONSTRAINT uk_system_oss_config_key UNIQUE (config_key);


--
-- Name: idx_built_in; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_built_in ON public.system_role USING btree (type);


--
-- Name: idx_error_log_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_error_log_created ON public.system_error_log USING btree (created_at DESC);


--
-- Name: idx_error_log_handle; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_error_log_handle ON public.system_error_log USING btree (handle_status);


--
-- Name: idx_error_log_level; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_error_log_level ON public.system_error_log USING btree (err_level);


--
-- Name: idx_error_log_trace; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_error_log_trace ON public.system_error_log USING btree (trace_id);


--
-- Name: idx_login_log_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_login_log_created ON public.system_login_log USING btree (created_at DESC);


--
-- Name: idx_login_log_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_login_log_name ON public.system_login_log USING btree (user_name);


--
-- Name: idx_login_log_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_login_log_status ON public.system_login_log USING btree (status);


--
-- Name: idx_oper_log_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_oper_log_created ON public.system_oper_log USING btree (created_at DESC);


--
-- Name: idx_oper_log_trace; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_oper_log_trace ON public.system_oper_log USING btree (trace_id);


--
-- Name: idx_oper_log_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_oper_log_type ON public.system_oper_log USING btree (oper_type);


--
-- Name: idx_oper_log_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_oper_log_user ON public.system_oper_log USING btree (user_id);


--
-- Name: idx_oss_config_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_oss_config_status ON public.system_oss_config USING btree (status) WHERE (del_flag = 0);


--
-- Name: idx_oss_rule_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_oss_rule_status ON public.system_oss_rule USING btree (status) WHERE (del_flag = 0);


--
-- Name: idx_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_role_id ON public.system_role_permission USING btree (role_id);


--
-- Name: idx_sys_message_config_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_message_config_type ON public.system_message_config USING btree (message_type, status);


--
-- Name: idx_sys_message_log_template_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_message_log_template_id ON public.system_message_log USING btree (message_template_id);


--
-- Name: idx_sys_message_log_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_message_log_time ON public.system_message_log USING btree (log_time DESC);


--
-- Name: idx_sys_message_template_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_message_template_key ON public.system_message_template USING btree (message_key, message_type, status);


--
-- Name: idx_sys_message_template_key_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_message_template_key_id ON public.system_message_template USING btree (message_key_id);


--
-- Name: idx_system_dept_parent; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_dept_parent ON public.system_dept USING btree (parent_id);


--
-- Name: idx_system_dict_data_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_dict_data_type ON public.system_dict_data USING btree (dict_type);


--
-- Name: idx_system_dict_type_group; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_dict_type_group ON public.system_dict_type USING btree (group_id);


--
-- Name: idx_system_job_run_job_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_job_run_job_time ON public.system_job_run USING btree (job_id, triggered_at DESC);


--
-- Name: idx_system_job_run_log_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_job_run_log_created_at ON public.system_job_run_log USING btree (created_at);


--
-- Name: idx_system_job_run_log_run_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_job_run_log_run_id ON public.system_job_run_log USING btree (run_id, id);


--
-- Name: idx_system_notice_read_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_notice_read_user ON public.system_notice_read USING btree (user_id);


--
-- Name: idx_system_notice_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_notice_status ON public.system_notice USING btree (status) WHERE (del_flag = 0);


--
-- Name: idx_system_notice_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_notice_type ON public.system_notice USING btree (notice_type) WHERE (del_flag = 0);


--
-- Name: idx_system_oss_create_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_oss_create_by ON public.system_oss USING btree (create_by, created_at) WHERE (del_flag = 0);


--
-- Name: idx_system_oss_original; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_oss_original ON public.system_oss USING btree (original_name) WHERE (del_flag = 0);


--
-- Name: idx_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_id ON public.system_user_role USING btree (user_id);


--
-- Name: mobile_phone; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX mobile_phone ON public."system_user" USING btree (mobile_phone);


--
-- Name: uk_oss_rule_domain_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_oss_rule_domain_name ON public.system_oss_rule USING btree (domain, rule_name) WHERE (del_flag = 0);


--
-- Name: uk_sys_message_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_sys_message_key ON public.system_message_key USING btree (message_key);


--
-- Name: uk_system_config_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_config_key ON public.system_config USING btree (config_key) WHERE (del_flag = 0);


--
-- Name: uk_system_dept_name_parent; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_dept_name_parent ON public.system_dept USING btree (dept_name, parent_id) WHERE (del_flag = 0);


--
-- Name: uk_system_dict_group_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_dict_group_name ON public.system_dict_group USING btree (group_name) WHERE (del_flag = 0);


--
-- Name: uk_system_dict_type_type; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_dict_type_type ON public.system_dict_type USING btree (dict_type) WHERE (del_flag = 0);


--
-- Name: uk_system_notice_read_user_notice; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_notice_read_user_notice ON public.system_notice_read USING btree (user_id, notice_id);


--
-- Name: uk_system_post_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_post_code ON public.system_post USING btree (post_code) WHERE (del_flag = 0);


--
-- Name: uk_system_role_dept; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_role_dept ON public.system_role_dept USING btree (role_id, dept_id);


--
-- Name: uk_system_role_menu; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_role_menu ON public.system_role_menu USING btree (role_id, menu_id);


--
-- Name: uk_system_user_post; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_system_user_post ON public.system_user_post USING btree (user_id, post_id);


--
-- Name: user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_id ON public."system_user" USING btree (id);


--
-- Name: username; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX username ON public."system_user" USING btree (username);


--
-- PostgreSQL database dump complete
--

\unrestrict h6vzkawMdZNbL8gFGgtHHg4bP3CY5jLXjjlkYCRaTr65qQ4RDij4rlMQ1k5jo0J

--
-- PostgreSQL database dump
--

\restrict hdgvfFMARchv8wieNJac8Q3D1Q681py9PO6OnS2br8yZRh8E7JTbxBrfw9oqazy

-- Dumped from database version 16.15
-- Dumped by pg_dump version 16.15

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: system_config; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_config VALUES (2, '分页默认大小', 'sys.page.size', '10', 'system', '列表默认 pageSize', 0, 0, '0001-01-01 00:00:00', '2026-10-08 09:53:07.929981');
INSERT INTO public.system_config VALUES (1, '系统名称', 'sys.app.name', 'Gosilite', 'global', '站点名称', 0, 0, '0001-01-01 00:00:00', '2026-10-09 17:13:58.191495');
INSERT INTO public.system_config VALUES (3, '登录最大失败次数', 'sys.login.max_failures', '5', 'system', '15分钟内达到该次数后临时锁定', 0, 0, '2026-10-13 07:36:28.34156', '2026-10-13 07:36:28.34156');
INSERT INTO public.system_config VALUES (4, '登录失败统计分钟', 'sys.login.failure_window_minutes', '15', 'system', '登录失败次数的统计窗口', 0, 0, '2026-10-13 07:36:28.34156', '2026-10-13 07:36:28.34156');
INSERT INTO public.system_config VALUES (5, '登录临时锁定分钟', 'sys.login.lock_minutes', '30', 'system', '达到失败阈值后的临时锁定时长', 0, 0, '2026-10-13 07:36:28.34156', '2026-10-13 07:36:28.34156');


--
-- Data for Name: system_dept; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_dept VALUES (100, 0, '0', '总公司', 1, 0, 0, '2026-10-05 07:18:38.333474', '2026-10-06 16:41:57.538725', 0);
INSERT INTO public.system_dept VALUES (102, 100, '0,100', '市场部门', 2, 0, 0, '2026-10-05 07:18:38.333474', '2026-10-06 17:23:41.156535', 0);
INSERT INTO public.system_dept VALUES (2107458799911571456, 0, '0', '测试，测试', 2, 0, 2, '2026-10-06 21:11:47.398355', '2026-10-06 21:11:57.186485', 0);
INSERT INTO public.system_dept VALUES (2107458891318038528, 0, '0', '测试，测试', 2, 0, 2, '2026-10-06 21:12:09.191288', '2026-10-06 21:12:16.961019', 0);
INSERT INTO public.system_dept VALUES (101, 100, '0,100', '研发部门', 2, 0, 0, '2026-10-05 07:18:38.333474', '2026-10-08 13:47:08.525868', 51);


--
-- Data for Name: system_dict_data; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_dict_data VALUES (4, 'sys_user_sex', '女', '1', '', 2, 0, 0, 0, '2026-10-05 08:46:55.544538', '2026-10-05 08:46:55.544538', 'default', '');
INSERT INTO public.system_dict_data VALUES (6, 'sys_state', '停用', 'disabled', '', 2, 0, 0, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'default', '');
INSERT INTO public.system_dict_data VALUES (13, 'sys_menu_type', '菜单', 'C', '', 2, 0, 0, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'default', '');
INSERT INTO public.system_dict_data VALUES (15, 'sys_config_type', '系统', 'system', '', 1, 0, 1, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'default', '');
INSERT INTO public.system_dict_data VALUES (41, 'sys_oper_status', '成功', '0', '', 1, 0, 1, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (17, 'sys_config_type', '全局', 'global', '', 3, 0, 0, 0, '2026-10-06 05:37:39.04561', '2026-10-06 05:37:39.04561', 'default', '');
INSERT INTO public.system_dict_data VALUES (43, 'sys_error_level', '普通', '1', '', 1, 0, 1, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (2, 'sys_status', '停用', '1', '', 2, 0, 0, 0, '2026-10-05 08:46:55.544538', '2026-10-05 08:46:55.544538', 'danger', '');
INSERT INTO public.system_dict_data VALUES (5, 'sys_state', '启用', 'enabled', '', 1, 0, 1, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'success', '');
INSERT INTO public.system_dict_data VALUES (12, 'sys_menu_type', '目录', 'M', '', 1, 0, 1, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'primary', '');
INSERT INTO public.system_dict_data VALUES (14, 'sys_menu_type', '按钮', 'F', '', 3, 0, 0, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'warning', '');
INSERT INTO public.system_dict_data VALUES (7, 'sys_data_scope', '全部数据', '1', '', 1, 0, 1, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'primary', '');
INSERT INTO public.system_dict_data VALUES (8, 'sys_data_scope', '本部门', '2', '', 2, 0, 0, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'primary', '');
INSERT INTO public.system_dict_data VALUES (9, 'sys_data_scope', '本部门及以下', '3', '', 3, 0, 0, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'primary', '');
INSERT INTO public.system_dict_data VALUES (10, 'sys_data_scope', '仅本人', '4', '', 4, 0, 0, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'primary', '');
INSERT INTO public.system_dict_data VALUES (11, 'sys_data_scope', '自定义部门', '5', '', 5, 0, 0, 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 'primary', '');
INSERT INTO public.system_dict_data VALUES (2107771497933180928, 'sys_status', '验收项', 'verify_9', '', 99, 0, 0, 2, '2026-10-07 17:54:20.265538', '2026-10-07 17:54:20.302503', 'warning', '#FF6A00');
INSERT INTO public.system_dict_data VALUES (44, 'sys_error_level', '严重', '2', '', 2, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (45, 'sys_handle_status', '待处理', '0', '', 1, 0, 1, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (46, 'sys_handle_status', '已处理', '1', '', 2, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (47, 'sys_login_status', '成功', '0', '', 1, 0, 1, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (48, 'sys_login_status', '失败', '1', '', 2, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (49, 'sys_login_status', '登出', '2', '', 3, 0, 0, 0, '2026-10-08 08:30:12.886662', '2026-10-08 08:30:12.886662', '', '');
INSERT INTO public.system_dict_data VALUES (42, 'sys_oper_status', '失败', '1', '', 2, 0, 0, 0, '0001-01-01 00:00:00', '2026-10-09 12:41:04.539629', 'danger', '');
INSERT INTO public.system_dict_data VALUES (50, 'sys_message_type', '邮件', 'MAIL', '', 1, 0, 1, 0, '2026-10-11 13:52:29.198316', '2026-10-11 13:52:29.198316', '', '');
INSERT INTO public.system_dict_data VALUES (51, 'sys_message_type', '短信', 'SMS', '', 2, 0, 0, 0, '2026-10-11 13:52:29.198316', '2026-10-11 13:52:29.198316', '', '');
INSERT INTO public.system_dict_data VALUES (52, 'sys_message_template_mode', '模板ID', 'TEMPLATE_ID', '', 1, 0, 0, 0, '2026-10-11 13:52:29.198316', '2026-10-11 13:52:29.198316', '', '');
INSERT INTO public.system_dict_data VALUES (16, 'sys_config_type', '业务', 'business', '', 2, 0, 0, 0, '0001-01-01 00:00:00', '2026-10-07 18:06:09.347546', 'default', '');
INSERT INTO public.system_dict_data VALUES (3, 'sys_user_sex', '男', '0', '', 1, 0, 1, 0, '0001-01-01 00:00:00', '2026-10-07 18:15:15.591734', '', '#3861ff');
INSERT INTO public.system_dict_data VALUES (2107777048217915392, 'sys_status', 't', 't_color', '', 99, 0, 0, 2, '2026-10-07 18:16:23.556903', '2026-10-07 18:16:23.592248', 'default', '#00FF88');
INSERT INTO public.system_dict_data VALUES (2107777048092086272, 'sys_status', 't', 't_info', '', 99, 0, 0, 2, '2026-10-07 18:16:23.526475', '2026-10-07 18:16:23.625724', 'info', '');
INSERT INTO public.system_dict_data VALUES (53, 'sys_message_template_mode', '模板内容', 'TEMPLATE_CONTENT', '', 2, 0, 1, 0, '2026-10-11 13:52:29.198316', '2026-10-11 13:52:29.198316', '', '');
INSERT INTO public.system_dict_data VALUES (2107777400317153280, 'sys_status', 't', 't_color2', '', 99, 0, 0, 2, '0001-01-01 00:00:00', '2026-10-07 18:17:47.557672', '', '#00FF88');
INSERT INTO public.system_dict_data VALUES (2107777400602365952, 'sys_status', 't', 't_info2', '', 99, 0, 0, 2, '2026-10-07 18:17:47.571496', '2026-10-07 18:17:47.61702', 'info', '');
INSERT INTO public.system_dict_data VALUES (1, 'sys_status', '启用', '0', '', 1, 0, 1, 0, '0001-01-01 00:00:00', '2026-10-08 08:29:43.255845', 'primary', '');
INSERT INTO public.system_dict_data VALUES (18, 'sys_action_type', '添加', 'add', '', 1, 0, 1, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (19, 'sys_action_type', '修改', 'edit', '', 2, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (20, 'sys_action_type', '删除', 'delete', '', 3, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (21, 'sys_action_type', '下载', 'download', '', 4, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (22, 'sys_action_type', '上传', 'upload', '', 5, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (23, 'sys_action_type', '查询', 'query', '', 6, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (24, 'sys_action_type', '导入', 'import', '', 7, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (25, 'sys_action_type', '导出', 'export', '', 8, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (26, 'sys_action_type', '审批', 'approve', '', 9, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (27, 'sys_action_type', '分配', 'assign', '', 10, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (28, 'sys_action_type', '其他', 'other', '', 99, 0, 0, 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', '', '');
INSERT INTO public.system_dict_data VALUES (29, 'sys_notice_type', '通知', '1', '', 1, 0, 1, 0, '2026-10-08 03:39:40.328222', '2026-10-08 03:39:40.328222', '', '');
INSERT INTO public.system_dict_data VALUES (30, 'sys_notice_type', '公告', '2', '', 2, 0, 0, 0, '2026-10-08 03:39:40.328222', '2026-10-08 03:39:40.328222', '', '');
INSERT INTO public.system_dict_data VALUES (31, 'sys_notice_status', '已发布', '0', '', 1, 0, 1, 0, '2026-10-08 03:39:40.328222', '2026-10-08 03:39:40.328222', '', '');
INSERT INTO public.system_dict_data VALUES (32, 'sys_notice_status', '已撤回', '1', '', 2, 0, 0, 0, '2026-10-08 03:39:40.328222', '2026-10-08 03:39:40.328222', '', '');
INSERT INTO public.system_dict_data VALUES (33, 'sys_oper_type', '其他', '0', '', 1, 0, 1, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (34, 'sys_oper_type', '新增', '1', '', 2, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (35, 'sys_oper_type', '修改', '2', '', 3, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (36, 'sys_oper_type', '删除', '3', '', 4, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (37, 'sys_oper_type', '查询', '4', '', 5, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (38, 'sys_oper_type', '导入', '5', '', 6, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (39, 'sys_oper_type', '导出', '6', '', 7, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');
INSERT INTO public.system_dict_data VALUES (40, 'sys_oper_type', '上传', '7', '', 8, 0, 0, 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', '', '');


--
-- Data for Name: system_dict_group; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_dict_group VALUES (2107655757913657344, '测试分组', 9, 'smoke', 2, '2026-10-07 10:14:25.695523', '2026-10-07 10:15:09.225689');
INSERT INTO public.system_dict_group VALUES (2107659177609400320, '前端验收组', 2, '', 2, '2026-10-07 10:28:01.018934', '2026-10-07 10:28:25.531904');
INSERT INTO public.system_dict_group VALUES (1, '系统', 1, '系统内置字典', 0, '0001-01-01 00:00:00', '2026-10-09 17:50:29.685191');
INSERT INTO public.system_dict_group VALUES (2107685992222691328, '业务测试', 2, '测试业务备注', 0, '0001-01-01 00:00:00', '2026-10-09 17:50:29.70431');


--
-- Data for Name: system_dict_type; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_dict_type VALUES (1, '状态', 'sys_status', 0, 1, '通用状态', 0, '2026-10-05 08:46:55.544538', '2026-10-05 08:46:55.544538', 1);
INSERT INTO public.system_dict_type VALUES (2, '性别', 'sys_user_sex', 0, 2, '用户性别', 0, '2026-10-05 08:46:55.544538', '2026-10-05 08:46:55.544538', 1);
INSERT INTO public.system_dict_type VALUES (3, '账号状态', 'sys_state', 0, 3, '启用/停用（值域 enabled/disabled）', 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 1);
INSERT INTO public.system_dict_type VALUES (4, '数据范围', 'sys_data_scope', 0, 4, '角色数据权限范围', 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 1);
INSERT INTO public.system_dict_type VALUES (5, '菜单类型', 'sys_menu_type', 0, 5, '目录/菜单/按钮', 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 1);
INSERT INTO public.system_dict_type VALUES (6, '参数类型', 'sys_config_type', 0, 6, '系统/业务参数', 0, '2026-10-06 03:34:43.621338', '2026-10-06 03:34:43.621338', 1);
INSERT INTO public.system_dict_type VALUES (2107655940181331968, '冒烟字典', 'smoke_status', 0, 1, '', 2, '2026-10-07 10:15:09.151252', '2026-10-07 10:15:09.211447', 2107655757913657344);
INSERT INTO public.system_dict_type VALUES (7, '按钮操作类型', 'sys_action_type', 0, 7, '按钮动作分类（操作审计用），F 菜单必选', 0, '2026-10-08 02:35:26.728581', '2026-10-08 02:35:26.728581', 1);
INSERT INTO public.system_dict_type VALUES (8, '通知公告类型', 'sys_notice_type', 0, 8, '通知公告分类（1 通知 2 公告）', 0, '2026-10-08 03:39:40.328222', '2026-10-08 03:39:40.328222', 1);
INSERT INTO public.system_dict_type VALUES (9, '通知公告状态', 'sys_notice_status', 0, 9, '发布状态（0 已发布 1 已撤回）', 0, '2026-10-08 03:39:40.328222', '2026-10-08 03:39:40.328222', 1);
INSERT INTO public.system_dict_type VALUES (10, '操作日志类型', 'sys_oper_type', 0, 10, '操作分类（0其他 1新增 2修改 3删除 4查询 5导入 6导出 7上传）', 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', 1);
INSERT INTO public.system_dict_type VALUES (11, '操作执行状态', 'sys_oper_status', 0, 11, '执行状态（0 成功 1 失败）', 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', 1);
INSERT INTO public.system_dict_type VALUES (12, '异常级别', 'sys_error_level', 0, 12, '异常严重级别（1 普通 2 严重）', 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', 1);
INSERT INTO public.system_dict_type VALUES (13, '异常处理状态', 'sys_handle_status', 0, 13, '异常处理状态（0 待处理 1 已处理）', 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', 1);
INSERT INTO public.system_dict_type VALUES (14, '登录状态', 'sys_login_status', 0, 14, '登录结果（0 成功 1 失败）', 0, '2026-10-08 07:06:16.504311', '2026-10-08 07:06:16.504311', 1);
INSERT INTO public.system_dict_type VALUES (2107030977934004224, '测试', 'sys_demo', 0, 9, 't', 2, '2026-10-05 16:51:46.539872', '2026-10-09 17:50:39.550939', 1);
INSERT INTO public.system_dict_type VALUES (15, '消息类型', 'sys_message_type', 0, 20, '消息通道类型（MAIL 邮件 SMS 短信）', 0, '2026-10-11 13:52:29.198316', '2026-10-11 13:52:29.198316', 1);
INSERT INTO public.system_dict_type VALUES (16, '消息模板模式', 'sys_message_template_mode', 0, 21, '模板ID模式 / 模板内容模式', 0, '2026-10-11 13:52:29.198316', '2026-10-11 13:52:29.198316', 1);


--
-- Data for Name: system_menu; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_menu VALUES (42, 3, 'system:role:list', '', 'C', '', '', '角色管理', 'role', '', 7, 'user-safety', '', '', '', 'system/role/index', 'show', '', 0, '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '', '', false, false, 'system:role:list', 0, '');
INSERT INTO public.system_menu VALUES (44, 3, 'system:user:list', '', 'C', '', '', '组织架构', 'contacts', '', 5, 'view-organization', '', '', '', 'system/organization/index', 'show', '', 0, '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '', '', false, false, 'system:user:list', 0, '');
INSERT INTO public.system_menu VALUES (43, 3, 'system:config:list', NULL, 'C', NULL, NULL, '参数管理', 'config', NULL, 8, 'setting', 'enabled', NULL, NULL, 'system/config/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:config:list', 0, '');
INSERT INTO public.system_menu VALUES (41, 3, 'system:post:list', NULL, 'C', NULL, NULL, '岗位管理', 'post', NULL, 6, 'work', 'enabled', NULL, NULL, 'system/post/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:post:list', 0, '');
INSERT INTO public.system_menu VALUES (4101, 41, 'system:post:add', NULL, 'F', NULL, NULL, '岗位新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:post:add', 0, 'add');
INSERT INTO public.system_menu VALUES (10, 3, 'page:root:page', NULL, 'C', NULL, 'layout', '菜单管理', 'menu', NULL, 4, 'menu', 'enabled', NULL, NULL, 'system/menu/index', 'show', 'enabled', 0, '2025-04-06 10:31:40', '2025-04-06 10:31:40', 'tokgos', 'tokgos', '', false, false, 'system:menu:list', 0, '');
INSERT INTO public.system_menu VALUES (2, 1, 'page:dashboard:index', NULL, 'C', NULL, 'layout', '概览仪表盘', 'base', NULL, 1, 'dashboard', 'enabled', NULL, NULL, 'dashboard/base/index', 'show', 'enabled', 0, '2025-03-31 16:38:37', '2025-03-31 16:38:37', 'tokgos', 'tokgos', '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (1, 0, 'page:root:dashboard', NULL, 'M', 'Worksapce', 'layout', '仪表盘', '/dashboard', NULL, 2, 'home', 'enabled', NULL, NULL, 'layout', 'show', 'enabled', 0, '2025-03-31 16:38:38', '2025-03-31 16:38:38', 'tokgos', 'tokgos', '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (3, 0, 'page:root:system', '1', 'M', '系统维护管理', 'layout', '系统管理', '/system', NULL, 2, 'setting', 'enabled', NULL, NULL, 'layout', 'show', 'enabled', 0, '2025-04-06 10:31:40', '2025-04-06 10:31:40', 'tokgos', 'tokgos', '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (9, 3, 'system:dict:list', '', 'C', '', '', '字典管理', 'dict', '', 7, 'book', '', '', '', 'system/dict/index', 'show', '', 0, '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '', '', false, false, 'system:dict:list', 0, '');
INSERT INTO public.system_menu VALUES (4102, 41, 'system:post:edit', NULL, 'F', NULL, NULL, '岗位修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:post:edit', 0, 'edit');
INSERT INTO public.system_menu VALUES (4103, 41, 'system:post:remove', NULL, 'F', NULL, NULL, '岗位删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:post:remove', 0, 'delete');
INSERT INTO public.system_menu VALUES (4201, 42, 'system:role:add', NULL, 'F', NULL, NULL, '角色新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:role:add', 0, 'add');
INSERT INTO public.system_menu VALUES (4202, 42, 'system:role:edit', NULL, 'F', NULL, NULL, '角色修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:role:edit', 0, 'edit');
INSERT INTO public.system_menu VALUES (4203, 42, 'system:role:remove', NULL, 'F', NULL, NULL, '角色删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:role:remove', 0, 'delete');
INSERT INTO public.system_menu VALUES (4204, 42, 'system:role:assign', NULL, 'F', NULL, NULL, '分配菜单', '', NULL, 4, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:role:assign', 0, 'assign');
INSERT INTO public.system_menu VALUES (4301, 43, 'system:config:add', NULL, 'F', NULL, NULL, '参数新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:config:add', 0, 'add');
INSERT INTO public.system_menu VALUES (4302, 43, 'system:config:edit', NULL, 'F', NULL, NULL, '参数修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:config:edit', 0, 'edit');
INSERT INTO public.system_menu VALUES (1001, 10, 'system:menu:add', NULL, 'F', NULL, NULL, '菜单新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:menu:add', 0, 'add');
INSERT INTO public.system_menu VALUES (1002, 10, 'system:menu:edit', NULL, 'F', NULL, NULL, '菜单修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:menu:edit', 0, 'edit');
INSERT INTO public.system_menu VALUES (1003, 10, 'system:menu:remove', NULL, 'F', NULL, NULL, '菜单删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:menu:remove', 0, 'delete');
INSERT INTO public.system_menu VALUES (9001, 9, 'system:dict:add', NULL, 'F', NULL, NULL, '字典新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:dict:add', 0, 'add');
INSERT INTO public.system_menu VALUES (9002, 9, 'system:dict:edit', NULL, 'F', NULL, NULL, '字典修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:dict:edit', 0, 'edit');
INSERT INTO public.system_menu VALUES (9003, 9, 'system:dict:remove', NULL, 'F', NULL, NULL, '字典删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:dict:remove', 0, 'delete');
INSERT INTO public.system_menu VALUES (4303, 43, 'system:config:remove', NULL, 'F', NULL, NULL, '参数删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:config:remove', 0, 'delete');
INSERT INTO public.system_menu VALUES (1201, 44, 'system:user:add', NULL, 'F', NULL, NULL, '用户新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:user:add', 0, 'add');
INSERT INTO public.system_menu VALUES (1202, 44, 'system:user:edit', NULL, 'F', NULL, NULL, '用户修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:user:edit', 0, 'edit');
INSERT INTO public.system_menu VALUES (4001, 44, 'system:dept:add', NULL, 'F', NULL, NULL, '部门新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:dept:add', 0, 'add');
INSERT INTO public.system_menu VALUES (4002, 44, 'system:dept:edit', NULL, 'F', NULL, NULL, '部门修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:dept:edit', 0, 'edit');
INSERT INTO public.system_menu VALUES (4003, 44, 'system:dept:remove', NULL, 'F', NULL, NULL, '部门删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:dept:remove', 0, 'delete');
INSERT INTO public.system_menu VALUES (1203, 44, 'system:user:remove', NULL, 'F', NULL, NULL, '用户删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:user:remove', 0, 'delete');
INSERT INTO public.system_menu VALUES (45, 3, 'system:notice:list', NULL, 'C', NULL, NULL, '通知公告', 'notice', NULL, 9, 'notification', 'enabled', NULL, NULL, 'system/notice/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:notice:list', 0, '');
INSERT INTO public.system_menu VALUES (4501, 45, 'system:notice:add', NULL, 'F', NULL, NULL, '通知新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:notice:add', 0, '');
INSERT INTO public.system_menu VALUES (4502, 45, 'system:notice:edit', NULL, 'F', NULL, NULL, '通知修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:notice:edit', 0, '');
INSERT INTO public.system_menu VALUES (4503, 45, 'system:notice:remove', NULL, 'F', NULL, NULL, '通知删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:notice:remove', 0, '');
INSERT INTO public.system_menu VALUES (4602, 46, 'system:loginlog:list', NULL, 'C', NULL, NULL, '登录日志', 'loginlog', NULL, 2, 'login', 'enabled', NULL, NULL, 'system/log/loginlog/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:loginlog:list', 0, '');
INSERT INTO public.system_menu VALUES (460101, 4601, 'system:operlog:clean', NULL, 'F', NULL, NULL, '清空日志', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:operlog:clean', 0, '');
INSERT INTO public.system_menu VALUES (460102, 4601, 'system:operlog:handle', NULL, 'F', NULL, NULL, '异常处理', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:operlog:handle', 0, '');
INSERT INTO public.system_menu VALUES (4901, 49, 'page:root:tool:docs', NULL, 'C', NULL, NULL, '开发文档', 'docs', NULL, 2, 'book-open', 'enabled', NULL, NULL, 'tool/docs/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (4601, 46, 'system:operlog:list', '', 'C', '', '', '操作日志', 'operlog', '', 1, 'catalog-1', '', '', '', 'system/log/operlog/index', 'show', '', 0, '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '', '', false, false, 'system:operlog:list', 0, '');
INSERT INTO public.system_menu VALUES (47, 0, 'page:root:monitor', NULL, 'M', NULL, NULL, '系统监控', '/monitor', NULL, 3, 'precise-monitor', 'enabled', NULL, NULL, 'layout', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (4701, 47, 'system:online:list', NULL, 'C', NULL, NULL, '在线用户', 'online', NULL, 1, 'user-safety', 'enabled', NULL, NULL, 'monitor/online/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:online:list', 0, '');
INSERT INTO public.system_menu VALUES (470101, 4701, 'system:online:force', NULL, 'F', NULL, NULL, '强退', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:online:force', 0, '');
INSERT INTO public.system_menu VALUES (4702, 47, 'system:monitor:list', NULL, 'C', NULL, NULL, '数据监控', 'data', NULL, 2, 'chart-line', 'enabled', NULL, NULL, 'monitor/data/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:monitor:list', 0, '');
INSERT INTO public.system_menu VALUES (470301, 4703, 'system:cachemonitor:delete', NULL, 'F', NULL, NULL, '删除', '', NULL, 1, 'delete', 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:cachemonitor:delete', 0, '');
INSERT INTO public.system_menu VALUES (4703, 47, 'system:cachemonitor:list', NULL, 'C', NULL, NULL, '缓存监控', 'cache', NULL, 3, 'layers', 'enabled', NULL, NULL, 'monitor/cache/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:cachemonitor:list', 0, '');
INSERT INTO public.system_menu VALUES (46, 3, '', '', 'M', '', '', '日志管理', '/log', '', 10, 'catalog', '', '', '', 'layout', 'show', '', 0, '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '', '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (60, 3, 'page:root:storage', NULL, 'M', NULL, NULL, '对象存储', 'storage', NULL, 11, 'layers', 'enabled', NULL, NULL, 'layout', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (61, 60, 'system:ossConfig:list', NULL, 'C', NULL, NULL, 'OSS配置', 'ossConfig', NULL, 1, 'file-code', 'enabled', NULL, NULL, 'system/ossConfig/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossConfig:list', 0, '');
INSERT INTO public.system_menu VALUES (62, 60, 'system:oss:list', NULL, 'C', NULL, NULL, '文件管理', 'oss', NULL, 2, 'folder', 'enabled', NULL, NULL, 'system/oss/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:oss:list', 0, '');
INSERT INTO public.system_menu VALUES (63, 60, 'system:ossRule:list', NULL, 'C', NULL, NULL, 'OSS处理规则', 'ossRule', NULL, 3, 'catalog', 'enabled', NULL, NULL, 'system/ossRule/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossRule:list', 0, '');
INSERT INTO public.system_menu VALUES (6104, 61, 'system:ossConfig:test', NULL, 'F', NULL, NULL, '配置测试', '', NULL, 4, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossConfig:test', 0, '');
INSERT INTO public.system_menu VALUES (6102, 61, 'system:ossConfig:edit', NULL, 'F', NULL, NULL, '配置修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossConfig:edit', 0, '');
INSERT INTO public.system_menu VALUES (6103, 61, 'system:ossConfig:remove', NULL, 'F', NULL, NULL, '配置删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossConfig:remove', 0, '');
INSERT INTO public.system_menu VALUES (6101, 61, 'system:ossConfig:add', NULL, 'F', NULL, NULL, '配置新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossConfig:add', 0, '');
INSERT INTO public.system_menu VALUES (6205, 62, 'system:oss:lock', NULL, 'F', NULL, NULL, '文件加锁解锁', '', NULL, 5, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:oss:lock', 0, '');
INSERT INTO public.system_menu VALUES (6203, 62, 'system:oss:edit', NULL, 'F', NULL, NULL, '文件修改', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:oss:edit', 0, '');
INSERT INTO public.system_menu VALUES (6204, 62, 'system:oss:remove', NULL, 'F', NULL, NULL, '文件删除', '', NULL, 4, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:oss:remove', 0, '');
INSERT INTO public.system_menu VALUES (6302, 63, 'system:ossRule:edit', NULL, 'F', NULL, NULL, '规则修改', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossRule:edit', 0, '');
INSERT INTO public.system_menu VALUES (6303, 63, 'system:ossRule:remove', NULL, 'F', NULL, NULL, '规则删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossRule:remove', 0, '');
INSERT INTO public.system_menu VALUES (6304, 63, 'system:ossRule:refresh', NULL, 'F', NULL, NULL, '刷新缓存', '', NULL, 4, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossRule:refresh', 0, '');
INSERT INTO public.system_menu VALUES (6301, 63, 'system:ossRule:add', NULL, 'F', NULL, NULL, '规则新增', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:ossRule:add', 0, '');
INSERT INTO public.system_menu VALUES (49, 0, 'page:root:tool', NULL, 'M', NULL, NULL, '系统工具', '/tool', NULL, 4, 'tools', 'enabled', NULL, NULL, 'layout', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (4704, 49, 'system:apimonitor:list', NULL, 'C', NULL, NULL, '系统接口', 'api', NULL, 1, 'file-code', 'enabled', NULL, NULL, 'tool/api/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:apimonitor:list', 0, '');
INSERT INTO public.system_menu VALUES (50, 3, 'page:root:message', NULL, 'M', NULL, NULL, '消息管理', 'message', NULL, 12, 'chat', 'enabled', NULL, NULL, 'layout', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, '', 0, '');
INSERT INTO public.system_menu VALUES (5001, 50, 'system:messageConfig:list', NULL, 'C', NULL, NULL, '消息配置', 'config', NULL, 1, 'tools', 'enabled', NULL, NULL, 'system/message/config/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageConfig:list', 0, '');
INSERT INTO public.system_menu VALUES (5002, 50, 'system:messageKey:list', NULL, 'C', NULL, NULL, '消息常量', 'key', NULL, 2, 'root-list', 'enabled', NULL, NULL, 'system/message/key/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageKey:list', 0, '');
INSERT INTO public.system_menu VALUES (5003, 50, 'system:messageTemplate:list', NULL, 'C', NULL, NULL, '消息模板', 'template', NULL, 3, 'file', 'enabled', NULL, NULL, 'system/message/template/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageTemplate:list', 0, '');
INSERT INTO public.system_menu VALUES (500104, 5001, 'system:messageConfig:remove', NULL, 'F', NULL, NULL, '配置删除', '', NULL, 4, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageConfig:remove', 0, '');
INSERT INTO public.system_menu VALUES (500103, 5001, 'system:messageConfig:edit', NULL, 'F', NULL, NULL, '配置修改', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageConfig:edit', 0, '');
INSERT INTO public.system_menu VALUES (500102, 5001, 'system:messageConfig:add', NULL, 'F', NULL, NULL, '配置新增', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageConfig:add', 0, '');
INSERT INTO public.system_menu VALUES (500101, 5001, 'system:messageConfig:query', NULL, 'F', NULL, NULL, '配置查询', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageConfig:query', 0, '');
INSERT INTO public.system_menu VALUES (500204, 5002, 'system:messageKey:remove', NULL, 'F', NULL, NULL, '常量删除', '', NULL, 4, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageKey:remove', 0, '');
INSERT INTO public.system_menu VALUES (500203, 5002, 'system:messageKey:edit', NULL, 'F', NULL, NULL, '常量修改', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageKey:edit', 0, '');
INSERT INTO public.system_menu VALUES (500202, 5002, 'system:messageKey:add', NULL, 'F', NULL, NULL, '常量新增', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageKey:add', 0, '');
INSERT INTO public.system_menu VALUES (500201, 5002, 'system:messageKey:query', NULL, 'F', NULL, NULL, '常量查询', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageKey:query', 0, '');
INSERT INTO public.system_menu VALUES (500304, 5003, 'system:messageTemplate:remove', NULL, 'F', NULL, NULL, '模板删除', '', NULL, 4, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageTemplate:remove', 0, '');
INSERT INTO public.system_menu VALUES (5004, 50, 'system:messageLog:list', NULL, 'C', NULL, NULL, '发送记录', 'log', NULL, 4, 'history', 'enabled', NULL, NULL, 'system/message/log/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageLog:list', 0, '');
INSERT INTO public.system_menu VALUES (500303, 5003, 'system:messageTemplate:edit', NULL, 'F', NULL, NULL, '模板修改', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageTemplate:edit', 0, '');
INSERT INTO public.system_menu VALUES (500302, 5003, 'system:messageTemplate:add', NULL, 'F', NULL, NULL, '模板新增', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageTemplate:add', 0, '');
INSERT INTO public.system_menu VALUES (500301, 5003, 'system:messageTemplate:query', NULL, 'F', NULL, NULL, '模板查询', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageTemplate:query', 0, '');
INSERT INTO public.system_menu VALUES (500402, 5004, 'system:messageLog:clear', NULL, 'F', NULL, NULL, '记录清空', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageLog:clear', 0, '');
INSERT INTO public.system_menu VALUES (500401, 5004, 'system:messageLog:query', NULL, 'F', NULL, NULL, '记录查询', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageLog:query', 0, '');
INSERT INTO public.system_menu VALUES (500403, 5004, 'system:messageLog:remove', NULL, 'F', NULL, NULL, '记录删除', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:messageLog:remove', 0, '');
INSERT INTO public.system_menu VALUES (5102, 51, 'system:scheduler:add', NULL, 'F', NULL, NULL, '任务新增', '', NULL, 2, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:scheduler:add', 0, '');
INSERT INTO public.system_menu VALUES (5104, 51, 'system:scheduler:remove', NULL, 'F', NULL, NULL, '任务删除', '', NULL, 4, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:scheduler:remove', 0, '');
INSERT INTO public.system_menu VALUES (5101, 51, 'system:scheduler:list', NULL, 'F', NULL, NULL, '任务查询', '', NULL, 1, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:scheduler:list', 0, '');
INSERT INTO public.system_menu VALUES (5105, 51, 'system:scheduler:run', NULL, 'F', NULL, NULL, '任务运行', '', NULL, 5, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:scheduler:run', 0, '');
INSERT INTO public.system_menu VALUES (5103, 51, 'system:scheduler:edit', NULL, 'F', NULL, NULL, '任务修改', '', NULL, 3, NULL, 'enabled', NULL, NULL, NULL, 'hide', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:scheduler:edit', 0, '');
INSERT INTO public.system_menu VALUES (51, 47, 'system:scheduler:list', NULL, 'C', NULL, NULL, '定时任务', 'scheduler', NULL, 5, 'time', 'enabled', NULL, NULL, 'system/scheduler/index', 'show', NULL, NULL, NULL, NULL, NULL, NULL, '', false, false, 'system:scheduler:list', 0, '');


--
-- Data for Name: system_oss_config; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_oss_config VALUES (1, 'local', 'gosilite', 'gosilite-secret', 'skyx', '', 'http://127.0.0.1:19000', '127.0.0.1:19000', 'N', '', '0', '1', '', '本地 RustFS 模板：上线后填写密钥并在页面「设为默认」生效', 0, '2026-10-10 11:59:49.458127', '2026-10-11 08:47:59.997335');


--
-- Data for Name: system_oss_rule; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_oss_rule VALUES (2109083399548309504, '80x80', '127.0.0.1:19000', 'image,.png', '#{#url}?x-oss-process=image/resize,w_80', 'N', 'N', '1', 0, '', 2, '2026-10-11 08:47:21.982435', '2026-10-11 09:41:43.991595');
INSERT INTO public.system_oss_rule VALUES (2109083399820939264, 'default-img', '127.0.0.1:19000', 'image', '#{#url}?v=ow', 'Y', 'Y', '1', 1, '', 2, '2026-10-11 08:47:22.049295', '2026-10-11 09:41:43.991595');
INSERT INTO public.system_oss_rule VALUES (1, '180x180', 'aliyuncs.com', 'image', '#{#url}?x-oss-process=image/auto-orient,1/resize,m_lfit,w_180,h_180/quality,q_90', 'N', 'N', '1', 10, '阿里云 180 缩略（头像/列表小图）', 0, '2026-10-11 02:22:57.204526', '2026-10-11 02:22:57.204526');
INSERT INTO public.system_oss_rule VALUES (2, '800x800', 'aliyuncs.com', 'image', '#{#url}?x-oss-process=image/auto-orient,1/resize,m_lfit,w_800,h_800/quality,q_90', 'N', 'N', '1', 11, '阿里云 800 展示图', 0, '2026-10-11 02:22:57.204526', '2026-10-11 02:22:57.204526');
INSERT INTO public.system_oss_rule VALUES (3, '180x180', 'myqcloud.com', 'image', '#{#url}?imageMogr2/thumbnail/180x180/quality/90', 'N', 'N', '1', 20, '腾讯云 COS 180 缩略', 0, '2026-10-11 02:22:57.204526', '2026-10-11 02:22:57.204526');
INSERT INTO public.system_oss_rule VALUES (4, '800x800', 'myqcloud.com', 'image', '#{#url}?imageMogr2/thumbnail/800x800/quality/90', 'N', 'N', '1', 21, '腾讯云 COS 800 展示图', 0, '2026-10-11 02:22:57.204526', '2026-10-11 02:22:57.204526');
INSERT INTO public.system_oss_rule VALUES (5, '180x180', 'myhuaweicloud.com', 'image', '#{#url}?x-obs-process=image/resize,m_lfit,w_180,h_180/quality,q_90', 'N', 'N', '1', 30, '华为云 OBS 180 缩略', 0, '2026-10-11 02:22:57.204526', '2026-10-11 02:22:57.204526');
INSERT INTO public.system_oss_rule VALUES (6, '800x800', 'myhuaweicloud.com', 'image', '#{#url}?x-obs-process=image/resize,m_lfit,w_800,h_800/quality,q_90', 'N', 'N', '1', 31, '华为云 OBS 800 展示图', 0, '2026-10-11 02:22:57.204526', '2026-10-11 02:22:57.204526');
INSERT INTO public.system_oss_rule VALUES (7, 'local-demo', '127.0.0.1', 'image', '#{#url}?x-oss-process=image/resize,m_lfit,w_800,h_800/quality,q_90', 'N', 'N', '1', 90, '本地示例（默认停用）：原生 S3 不处理图片参数，带 query 会 AccessDenied；对接 CDN/图片网关后启用', 0, '2026-10-11 02:22:57.204526', '2026-10-11 10:44:25.448239');


--
-- Data for Name: system_post; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_post VALUES (1, 'CEO', '董事长', 1, 0, 0, '2026-10-05 07:18:38.333474', '2026-10-05 07:18:38.333474');
INSERT INTO public.system_post VALUES (2, 'CTO', '技术负责人', 2, 0, 0, '2026-10-05 07:18:38.333474', '2026-10-05 07:18:38.333474');
INSERT INTO public.system_post VALUES (3, 'PM', '项目经理', 3, 0, 0, '2026-10-05 07:18:38.333474', '2026-10-05 07:18:38.333474');
INSERT INTO public.system_post VALUES (2107012545083609088, 'DEV', '开发', 9, 0, 0, '2026-10-05 15:38:31.806204', '2026-10-05 15:38:31.806204');


--
-- Data for Name: system_role; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_role VALUES (1, '系统管理员', 'system', '系统管理员', 'enabled', 'built-y', 0, '2026-10-12 10:39:56.10032', '2025-03-31 16:39:35', 51, 51, NULL, 1, 0);


--
-- Data for Name: system_role_dept; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: system_role_menu; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_role_menu VALUES (83, 1, 51, '2026-10-12 05:06:02.712263');
INSERT INTO public.system_role_menu VALUES (84, 1, 5102, '2026-10-12 05:06:02.712263');
INSERT INTO public.system_role_menu VALUES (85, 1, 5104, '2026-10-12 05:06:02.712263');
INSERT INTO public.system_role_menu VALUES (86, 1, 5101, '2026-10-12 05:06:02.712263');
INSERT INTO public.system_role_menu VALUES (87, 1, 5105, '2026-10-12 05:06:02.712263');
INSERT INTO public.system_role_menu VALUES (88, 1, 5103, '2026-10-12 05:06:02.712263');
INSERT INTO public.system_role_menu VALUES (2109474116129001472, 1, 1, '2026-10-12 10:39:56.076036');
INSERT INTO public.system_role_menu VALUES (2109474116133195776, 1, 0, '2026-10-12 10:39:56.076871');
INSERT INTO public.system_role_menu VALUES (2109474116133195777, 1, 2, '2026-10-12 10:39:56.077195');
INSERT INTO public.system_role_menu VALUES (2109474116133195778, 1, 3, '2026-10-12 10:39:56.077534');
INSERT INTO public.system_role_menu VALUES (2109474116137390080, 1, 10, '2026-10-12 10:39:56.07783');
INSERT INTO public.system_role_menu VALUES (2109474116137390081, 1, 44, '2026-10-12 10:39:56.078165');
INSERT INTO public.system_role_menu VALUES (2109474116137390082, 1, 41, '2026-10-12 10:39:56.078513');
INSERT INTO public.system_role_menu VALUES (2109474116141584384, 1, 9, '2026-10-12 10:39:56.078806');
INSERT INTO public.system_role_menu VALUES (2109474116141584385, 1, 42, '2026-10-12 10:39:56.079155');
INSERT INTO public.system_role_menu VALUES (2109474116141584386, 1, 43, '2026-10-12 10:39:56.079525');
INSERT INTO public.system_role_menu VALUES (2109474116145778688, 1, 45, '2026-10-12 10:39:56.079846');
INSERT INTO public.system_role_menu VALUES (2109474116145778689, 1, 46, '2026-10-12 10:39:56.080181');
INSERT INTO public.system_role_menu VALUES (2109474116145778690, 1, 4601, '2026-10-12 10:39:56.080521');
INSERT INTO public.system_role_menu VALUES (2109474116149972992, 1, 4602, '2026-10-12 10:39:56.080792');
INSERT INTO public.system_role_menu VALUES (2109474116149972993, 1, 60, '2026-10-12 10:39:56.081106');
INSERT INTO public.system_role_menu VALUES (2109474116149972994, 1, 61, '2026-10-12 10:39:56.081369');
INSERT INTO public.system_role_menu VALUES (2109474116149972995, 1, 62, '2026-10-12 10:39:56.081647');
INSERT INTO public.system_role_menu VALUES (2109474116154167296, 1, 63, '2026-10-12 10:39:56.081931');
INSERT INTO public.system_role_menu VALUES (2109474116154167297, 1, 50, '2026-10-12 10:39:56.082226');
INSERT INTO public.system_role_menu VALUES (2109474116154167298, 1, 5001, '2026-10-12 10:39:56.082527');
INSERT INTO public.system_role_menu VALUES (2109474116158361600, 1, 5002, '2026-10-12 10:39:56.082858');
INSERT INTO public.system_role_menu VALUES (2109474116158361601, 1, 5003, '2026-10-12 10:39:56.083209');
INSERT INTO public.system_role_menu VALUES (2109474116158361602, 1, 5004, '2026-10-12 10:39:56.083508');
INSERT INTO public.system_role_menu VALUES (2109474116162555904, 1, 47, '2026-10-12 10:39:56.083753');
INSERT INTO public.system_role_menu VALUES (2109474116162555905, 1, 4701, '2026-10-12 10:39:56.08399');
INSERT INTO public.system_role_menu VALUES (2109474116162555906, 1, 4702, '2026-10-12 10:39:56.084307');
INSERT INTO public.system_role_menu VALUES (2109474116162555907, 1, 4703, '2026-10-12 10:39:56.084612');
INSERT INTO public.system_role_menu VALUES (2109474116166750208, 1, 49, '2026-10-12 10:39:56.084883');
INSERT INTO public.system_role_menu VALUES (2109474116166750209, 1, 4704, '2026-10-12 10:39:56.085119');
INSERT INTO public.system_role_menu VALUES (2109474116166750210, 1, 4901, '2026-10-12 10:39:56.085362');
INSERT INTO public.system_role_menu VALUES (2109474116166750211, 1, 1001, '2026-10-12 10:39:56.085577');
INSERT INTO public.system_role_menu VALUES (2109474116170944512, 1, 1002, '2026-10-12 10:39:56.085808');
INSERT INTO public.system_role_menu VALUES (2109474116170944513, 1, 1003, '2026-10-12 10:39:56.086024');
INSERT INTO public.system_role_menu VALUES (2109474116170944514, 1, 1201, '2026-10-12 10:39:56.086267');
INSERT INTO public.system_role_menu VALUES (2109474116170944515, 1, 4001, '2026-10-12 10:39:56.086498');
INSERT INTO public.system_role_menu VALUES (2109474116175138816, 1, 1202, '2026-10-12 10:39:56.086892');
INSERT INTO public.system_role_menu VALUES (2109474116175138817, 1, 4002, '2026-10-12 10:39:56.087151');
INSERT INTO public.system_role_menu VALUES (2109474116175138818, 1, 1203, '2026-10-12 10:39:56.087437');
INSERT INTO public.system_role_menu VALUES (2109474116179333120, 1, 4003, '2026-10-12 10:39:56.087713');
INSERT INTO public.system_role_menu VALUES (2109474116179333121, 1, 4101, '2026-10-12 10:39:56.087995');
INSERT INTO public.system_role_menu VALUES (2109474116179333122, 1, 4102, '2026-10-12 10:39:56.088271');
INSERT INTO public.system_role_menu VALUES (2109474116179333123, 1, 4103, '2026-10-12 10:39:56.088563');
INSERT INTO public.system_role_menu VALUES (2109474116183527424, 1, 9001, '2026-10-12 10:39:56.088829');
INSERT INTO public.system_role_menu VALUES (2109474116183527425, 1, 9002, '2026-10-12 10:39:56.089092');
INSERT INTO public.system_role_menu VALUES (2109474116183527426, 1, 9003, '2026-10-12 10:39:56.089458');
INSERT INTO public.system_role_menu VALUES (2109474116187721728, 1, 4201, '2026-10-12 10:39:56.089797');
INSERT INTO public.system_role_menu VALUES (2109474116187721729, 1, 4202, '2026-10-12 10:39:56.09009');
INSERT INTO public.system_role_menu VALUES (2109474116187721730, 1, 4203, '2026-10-12 10:39:56.090358');
INSERT INTO public.system_role_menu VALUES (2109474116191916032, 1, 4204, '2026-10-12 10:39:56.090692');
INSERT INTO public.system_role_menu VALUES (2109474116191916033, 1, 4301, '2026-10-12 10:39:56.090897');
INSERT INTO public.system_role_menu VALUES (2109474116191916034, 1, 4302, '2026-10-12 10:39:56.091084');
INSERT INTO public.system_role_menu VALUES (2109474116191916035, 1, 4303, '2026-10-12 10:39:56.091256');
INSERT INTO public.system_role_menu VALUES (2109474116191916036, 1, 4501, '2026-10-12 10:39:56.091438');
INSERT INTO public.system_role_menu VALUES (2109474116191916037, 1, 4502, '2026-10-12 10:39:56.09161');
INSERT INTO public.system_role_menu VALUES (2109474116196110336, 1, 4503, '2026-10-12 10:39:56.092129');
INSERT INTO public.system_role_menu VALUES (2109474116196110337, 1, 460101, '2026-10-12 10:39:56.09238');
INSERT INTO public.system_role_menu VALUES (2109474116196110338, 1, 460102, '2026-10-12 10:39:56.09261');
INSERT INTO public.system_role_menu VALUES (2109474116200304640, 1, 6101, '2026-10-12 10:39:56.092848');
INSERT INTO public.system_role_menu VALUES (2109474116200304641, 1, 6102, '2026-10-12 10:39:56.093075');
INSERT INTO public.system_role_menu VALUES (2109474116200304642, 1, 6103, '2026-10-12 10:39:56.093245');
INSERT INTO public.system_role_menu VALUES (2109474116200304643, 1, 6104, '2026-10-12 10:39:56.093408');
INSERT INTO public.system_role_menu VALUES (2109474116200304644, 1, 6203, '2026-10-12 10:39:56.093573');
INSERT INTO public.system_role_menu VALUES (2109474116204498944, 1, 6204, '2026-10-12 10:39:56.09375');
INSERT INTO public.system_role_menu VALUES (2109474116204498945, 1, 6205, '2026-10-12 10:39:56.093931');
INSERT INTO public.system_role_menu VALUES (2109474116204498946, 1, 6301, '2026-10-12 10:39:56.094108');
INSERT INTO public.system_role_menu VALUES (2109474116204498947, 1, 6302, '2026-10-12 10:39:56.094283');
INSERT INTO public.system_role_menu VALUES (2109474116204498948, 1, 6303, '2026-10-12 10:39:56.094465');
INSERT INTO public.system_role_menu VALUES (2109474116204498949, 1, 6304, '2026-10-12 10:39:56.094623');
INSERT INTO public.system_role_menu VALUES (2109474116208693248, 1, 500101, '2026-10-12 10:39:56.094805');
INSERT INTO public.system_role_menu VALUES (2109474116208693249, 1, 500102, '2026-10-12 10:39:56.094981');
INSERT INTO public.system_role_menu VALUES (2109474116208693250, 1, 500103, '2026-10-12 10:39:56.095158');
INSERT INTO public.system_role_menu VALUES (2109474116208693251, 1, 500104, '2026-10-12 10:39:56.095333');
INSERT INTO public.system_role_menu VALUES (2109474116208693252, 1, 500201, '2026-10-12 10:39:56.095529');
INSERT INTO public.system_role_menu VALUES (2109474116212887552, 1, 500202, '2026-10-12 10:39:56.095688');
INSERT INTO public.system_role_menu VALUES (2109474116212887553, 1, 500203, '2026-10-12 10:39:56.095848');
INSERT INTO public.system_role_menu VALUES (2109474116212887554, 1, 500204, '2026-10-12 10:39:56.096015');
INSERT INTO public.system_role_menu VALUES (2109474116212887555, 1, 500301, '2026-10-12 10:39:56.096168');
INSERT INTO public.system_role_menu VALUES (2109474116212887556, 1, 500302, '2026-10-12 10:39:56.096351');
INSERT INTO public.system_role_menu VALUES (2109474116212887557, 1, 500303, '2026-10-12 10:39:56.096503');
INSERT INTO public.system_role_menu VALUES (2109474116217081856, 1, 500304, '2026-10-12 10:39:56.096711');
INSERT INTO public.system_role_menu VALUES (2109474116217081857, 1, 500401, '2026-10-12 10:39:56.096881');
INSERT INTO public.system_role_menu VALUES (2109474116217081858, 1, 500402, '2026-10-12 10:39:56.097031');
INSERT INTO public.system_role_menu VALUES (2109474116217081859, 1, 500403, '2026-10-12 10:39:56.097214');
INSERT INTO public.system_role_menu VALUES (2109474116217081860, 1, 470101, '2026-10-12 10:39:56.097366');
INSERT INTO public.system_role_menu VALUES (2109474116217081861, 1, 470301, '2026-10-12 10:39:56.097492');


--
-- Data for Name: system_role_permission; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_role_permission VALUES (1875116722746462208, 1, 3, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722750656512, 1, 6, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722750656513, 1, 8, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722750656514, 1, 25, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722750656515, 1, 7, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722750656516, 1, 5, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722750656517, 1, 28, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722754850816, 1, 27, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722754850817, 1, 26, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722754850818, 1, 4, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722754850819, 1, 24, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722754850820, 1, 23, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722754850821, 1, 22, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722759045120, 1, 11, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722759045121, 1, 12, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722759045123, 1, 9, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722759045124, 1, 10, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722763239424, 1, 1, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722763239425, 1, 2, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722763239426, 1, 14, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722763239427, 1, 15, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722763239428, 1, 16, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722763239429, 1, 17, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722763239430, 1, 18, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722767433728, 1, 19, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875116722767433730, 1, 21, 0, '2025-01-03 17:47:19', '2025-01-03 17:47:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875221559326924800, 1, 29, 0, '2025-01-04 00:43:54', '2025-01-04 00:43:54', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1875221559343702016, 1, 30, 0, '2025-01-04 00:43:54', '2025-01-04 00:43:54', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1876157490659479552, 1, 32, 0, '2025-01-06 14:42:58', '2025-01-06 14:42:58', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1876157490663673856, 1, 31, 0, '2025-01-06 14:42:58', '2025-01-06 14:42:58', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1889341452091514880, 1, 1889120294364499968, 0, '2025-02-11 23:51:19', '2025-02-11 23:51:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1889341452091514881, 1, 1889120913158557696, 0, '2025-02-11 23:51:19', '2025-02-11 23:51:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1889341452091514882, 1, 1889120988593115136, 0, '2025-02-11 23:51:19', '2025-02-11 23:51:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1889341452099903488, 1, 1889122917503193088, 0, '2025-02-11 23:51:19', '2025-02-11 23:51:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1889341452099903489, 1, 1889143726988840960, 0, '2025-02-11 23:51:19', '2025-02-11 23:51:19', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1903360400145182720, 1, 1903358641263800320, 0, '2025-03-22 16:17:37', '2025-03-22 16:17:37', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1903360400145182721, 1, 1903358872265093120, 0, '2025-03-22 16:17:37', '2025-03-22 16:17:37', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1903360400149377024, 1, 1903359459782225920, 0, '2025-03-22 16:17:37', '2025-03-22 16:17:37', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1903368689704693760, 1, 1903368448997781504, 0, '2025-03-22 16:50:33', '2025-03-22 16:50:33', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1903368689708888064, 1, 1903368637762433024, 0, '2025-03-22 16:50:33', '2025-03-22 16:50:33', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1903382566853083136, 1, 1903367506814496768, 0, '2025-03-22 17:45:42', '2025-03-22 17:45:42', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948155244544, 1, 1904026472905109504, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948155244545, 1, 1905550015585304576, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948159438848, 1, 1905550106232602624, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948159438849, 1, 1905550183487488000, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948159438850, 1, 1905550319848505344, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948159438851, 1, 1905550529723088896, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948167827456, 1, 1905550598811664384, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948167827457, 1, 1905550684115419136, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948167827458, 1, 1905550750255398912, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1905550948167827459, 1, 1905550848125288448, 0, '2025-03-28 17:22:04', '2025-03-28 17:22:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1906035987113758720, 1, 1906035392764104704, 0, '2025-03-30 01:29:26', '2025-03-30 01:29:26', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1906035987122147328, 1, 1906035808730009600, 0, '2025-03-30 01:29:26', '2025-03-30 01:29:26', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1906035987122147329, 1, 1906035922336927744, 0, '2025-03-30 01:29:26', '2025-03-30 01:29:26', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1907730260383961088, 1, 1907730130071130112, 0, '2025-04-03 17:41:53', '2025-04-03 17:41:53', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1907730260388155392, 1, 1907730220076699648, 0, '2025-04-03 17:41:53', '2025-04-03 17:41:53', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902357291008, 1, 1908707315709038592, 0, '2025-04-06 10:34:37', '2025-04-06 10:34:37', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902369873920, 1, 1908707472336932864, 0, '2025-04-06 10:34:37', '2025-04-06 10:34:37', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902374068224, 1, 1908707543484911616, 0, '2025-04-06 10:34:37', '2025-04-06 10:34:37', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902378262528, 1, 1908707660606656512, 0, '2025-04-06 10:34:37', '2025-04-06 10:34:37', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902382456832, 1, 1908707746375979008, 0, '2025-04-06 10:34:38', '2025-04-06 10:34:38', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902390845440, 1, 1908707886834831360, 0, '2025-04-06 10:34:38', '2025-04-06 10:34:38', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902390845441, 1, 1908707979499589632, 0, '2025-04-06 10:34:38', '2025-04-06 10:34:38', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902395039744, 1, 1908708069584850944, 0, '2025-04-06 10:34:38', '2025-04-06 10:34:38', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1908709902395039745, 1, 1908708195812429824, 0, '2025-04-06 10:34:38', '2025-04-06 10:34:38', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1913847313659785216, 1, 1913847237470253056, 0, '2025-04-20 14:48:52', '2025-04-20 14:48:52', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1949408081893978112, 1, 1949408006492975104, 0, '2025-07-27 17:54:40', '2025-07-27 17:54:40', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1949409311248674816, 1, 1949409257930682368, 0, '2025-07-27 17:59:33', '2025-07-27 17:59:33', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1949423379040788480, 1, 1949423350762790912, 0, '2025-07-27 18:55:27', '2025-07-27 18:55:27', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1954473407206895616, 1, 1954473350973861888, 0, '2025-08-10 17:22:27', '2025-08-10 17:22:27', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1956260628272410624, 1, 1956260539218948096, 0, '2025-08-15 15:44:14', '2025-08-15 15:44:14', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1959420387641180160, 1, 1959420229826297856, 0, '2025-08-24 08:59:59', '2025-08-24 08:59:59', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1959879351742910464, 1, 1959879303399362560, 0, '2025-08-25 15:23:45', '2025-08-25 15:23:45', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1959917096221728768, 1, 1959917030211772416, 0, '2025-08-25 17:53:44', '2025-08-25 17:53:44', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1959918322975956993, 1, 1959918278843490304, 0, '2025-08-25 17:58:36', '2025-08-25 17:58:36', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1960652779089551360, 1, 1960652018251194368, 0, '2025-08-27 18:37:04', '2025-08-27 18:37:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1960652779093745664, 1, 1960652732666994688, 0, '2025-08-27 18:37:04', '2025-08-27 18:37:04', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1960654077469794304, 1, 1960654050164875264, 0, '2025-08-27 18:42:14', '2025-08-27 18:42:14', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1961025429053792256, 1, 1961025393704198144, 0, '2025-08-28 19:17:51', '2025-08-28 19:17:51', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1961593911944318976, 1, 1961593861084188672, 0, '2025-08-30 08:56:48', '2025-08-30 08:56:48', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1962395680823717888, 1, 1962395639136530432, 0, '2025-09-01 14:02:45', '2025-09-01 14:02:45', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1962706157269303296, 1, 1962706119839334400, 0, '2025-09-02 10:36:28', '2025-09-02 10:36:28', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619874443264, 1, 1963518937572552704, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619874443265, 1, 1963522053235785728, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619878637568, 1, 1963521737345974272, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619878637569, 1, 1963524054535319552, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619882831872, 1, 1963524769156640768, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619882831873, 1, 1963524293593870336, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619882831874, 1, 1963526479556091904, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619887026176, 1, 1963526579187589120, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619887026177, 1, 1963526650192961536, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619887026178, 1, 1963525958128607232, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619887026179, 1, 1963526090681196544, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1963551619887026180, 1, 1963526180602880000, 0, '2025-09-04 18:36:02', '2025-09-04 18:36:02', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965367309535178752, 1, 1965364173386969088, 0, '2025-09-09 18:50:56', '2025-09-09 18:50:56', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965367309539373056, 1, 1965365132846592000, 0, '2025-09-09 18:50:56', '2025-09-09 18:50:56', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965368609102524416, 1, 1965368470929567744, 0, '2025-09-09 18:56:06', '2025-09-09 18:56:06', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965368609106718720, 1, 1965368572083597312, 0, '2025-09-09 18:56:06', '2025-09-09 18:56:06', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965371736732123136, 1, 1965371530401726464, 0, '2025-09-09 19:08:32', '2025-09-09 19:08:32', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965371736732123137, 1, 1965371616246546432, 0, '2025-09-09 19:08:32', '2025-09-09 19:08:32', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965371736732123138, 1, 1965371690682859520, 0, '2025-09-09 19:08:32', '2025-09-09 19:08:32', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965675670780768256, 1, 1965675332237520896, 0, '2025-09-10 15:16:15', '2025-09-10 15:16:15', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1965675670784962560, 1, 1965675627373916160, 0, '2025-09-10 15:16:15', '2025-09-10 15:16:15', 'tokgos', 'tokgos');
INSERT INTO public.system_role_permission VALUES (1967880589822599168, 1, 1967880552946278400, 0, '2025-09-16 17:17:49', '2025-09-16 17:17:49', 'tokgos', 'tokgos');


--
-- Data for Name: system_user; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public."system_user" VALUES (51, 'skyx', '17637474818', 'tokgos@outlook.com', 'Super Admin', 'uploads/20261011/bf05a839-6160-49ba-a873-3f4cf4d1a494.png', 'enabled', '$2a$10$WFyk1Cq6ov7s0b2bAFDo9.BQPxe3vLqqmu85FVWPvTbXXsi0GryCK', 0, '2025-04-09 10:25:10', '2025-04-09 10:25:10', 'tokgos', 'tokgos', 101, 0, false);


--
-- Data for Name: system_user_acl_permission; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: system_user_post; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: system_user_role; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.system_user_role VALUES (2107644802639073280, 51, 1, '', 0, '0001-01-01 00:00:00', '0001-01-01 00:00:00', '', '');


--
-- Name: system_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_config_id_seq', 5, true);


--
-- Name: system_dept_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_dept_id_seq', 1, false);


--
-- Name: system_dict_data_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_dict_data_id_seq', 53, true);


--
-- Name: system_dict_group_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_dict_group_id_seq', 1, true);


--
-- Name: system_dict_type_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_dict_type_id_seq', 16, true);


--
-- Name: system_oss_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_oss_config_id_seq', 1, true);


--
-- Name: system_oss_rule_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_oss_rule_id_seq', 7, true);


--
-- Name: system_post_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_post_id_seq', 1, false);


--
-- Name: system_role_dept_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_role_dept_id_seq', 1, false);


--
-- Name: system_role_menu_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_role_menu_id_seq', 88, true);


--
-- Name: system_user_post_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.system_user_post_id_seq', 1, false);


--
-- PostgreSQL database dump complete
--

\unrestrict hdgvfFMARchv8wieNJac8Q3D1Q681py9PO6OnS2br8yZRh8E7JTbxBrfw9oqazy

