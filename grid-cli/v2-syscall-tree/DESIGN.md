# PromiseGrid Design

## Overview

This document outlines the design considerations for implementing PromiseGrid.

## Message Structure

- **Message Structure**: The `Message` structure will have a unified
  format where the first element in the `Parms` field is the promise.

## Module Interface

- **Single `Accept()` Function**: The Module interface will include a
  single `Accept()` function that returns a promise message. 

## Hierarchical Syscall Tree

- **Ant Routing Mechanism**: A hierarchical syscall tree functions
  as an "ant routing" mechanism, caching successful paths to optimize
  future routing. This dynamic routing approach efficiently manages
  message delivery based on cached paths.

## Dynamic Acceptance History is the Syscall Tree

- The syscall tree captures positive and negative acceptance history. 

## Promise-Based Design

- **Deep Integration of Promises**: The design emphasizes
  "promises all the way down," ensuring the `Accept()` response itself
  is a promise message. 


