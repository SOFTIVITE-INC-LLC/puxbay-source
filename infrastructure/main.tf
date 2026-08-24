terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "puxbay-vpc"
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "puxbay-cluster"
}

# RDS Database
resource "aws_db_instance" "postgres" {
  identifier           = "puxbay-db"
  engine               = "postgres"
  engine_version       = "15.3"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  db_name              = "puxbay_go"
  username             = var.db_username
  password             = var.db_password
  skip_final_snapshot  = true
}

# S3 Bucket for Backups
resource "aws_s3_bucket" "backups" {
  bucket = "puxbay-database-backups"
}
